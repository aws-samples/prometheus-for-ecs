package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
)

const (
	IpAddressAttribute      = "AWS_INSTANCE_IPV4"
	PortNumberAttribute     = "AWS_INSTANCE_PORT"
	ClusterNameAttribute    = "ECS_CLUSTER_NAME"
	ServiceNameAttribute    = "ECS_SERVICE_NAME"
	TaskDefinitionAttribute = "ECS_TASK_DEFINITION_FAMILY"
	MetricsPortTag          = "METRICS_PORT"
	MetricsPathTag          = "METRICS_PATH"
	MetricsSchemeTag        = "METRICS_SCHEME"
	EcsMetricsPortTag       = "ECS_METRICS_PORT"
	EcsMetricsPathTag       = "ECS_METRICS_PATH"
)

type CloudMapClient struct {
	service *servicediscovery.ServiceDiscovery
}

type ServiceDiscoveryInstance struct {
	service    *string
	instanceId *string
	attributes map[string]*string
}

type InstanceScrapeConfig struct {
	Targets []string          `json:"targets,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}

// Retrieve a JSON object that provides a list of ECS targets to be scraped for Prometheus metrics
func GetPrometheusScrapeConfig(selectedNamespaces []string) *string {
	client := &CloudMapClient{service: servicediscovery.New(sharedSession)}

	sdNamespaces, _ := client.getNamespaces()
	sdServicesMap, _ := client.getServices(selectedNamespaces, sdNamespaces)
	scrapeConfigurations := make([]*InstanceScrapeConfig, 0)
	for sdNamespace, sdServices := range sdServicesMap {
		for _, service := range sdServices {
			serviceTags := client.getServiceTags(service)
			sdInstances, _ := client.getInstances(service)
			for _, instance := range sdInstances {
				appScrapeConfig, _ := client.getInstanceScrapeConfigurationApplication(instance, serviceTags, &sdNamespace)
				infraScrapeConfig, _ := client.getInstanceScrapeConfigurationInfrastructure(instance, serviceTags, &sdNamespace)
				if appScrapeConfig != nil {
					scrapeConfigurations = append(scrapeConfigurations, appScrapeConfig)
				}
				if infraScrapeConfig != nil {
					scrapeConfigurations = append(scrapeConfigurations, infraScrapeConfig)
				}
			}
		}
	}

	jsonBytes, err := json.MarshalIndent(scrapeConfigurations, "", "   ")
	if err != nil {
		log.Println(err)
		return aws.String("")
	}
	jsonString := string(jsonBytes)
	return &jsonString
}

// Get a list of all ServiceDiscovery namespaces and their respective IDs available under Cloud Map
func (c *CloudMapClient) getNamespaces() (map[string]string, error) {
	filterType := aws.String("TYPE")
	filterCondition := aws.String("EQ")
	filterValues := []*string{aws.String("DNS_PRIVATE")}
	namespaceFilter := servicediscovery.NamespaceFilter{
		Name:      filterType,
		Values:    filterValues,
		Condition: filterCondition}
	listNamespacesOutput, err := c.service.ListNamespaces(&servicediscovery.ListNamespacesInput{Filters: []*servicediscovery.NamespaceFilter{&namespaceFilter}})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	sdNamespaces := make(map[string]string)
	for _, namespaceSummary := range listNamespacesOutput.Namespaces {
		sdNamespaces[*namespaceSummary.Name] = *namespaceSummary.Id
	}
	return sdNamespaces, nil
}

// Cycle through each ServiceDiscovery namespace and find the list of ServiceDiscovery services
func (c *CloudMapClient) getServices(selectedNamespaces []string, sdNamespaces map[string]string) (map[string][]*servicediscovery.ServiceSummary, error) {
	sdServicesMap := make(map[string][]*servicediscovery.ServiceSummary)
	sdServicesCount := 0
	for _, name := range selectedNamespaces {
		sdServices := make([]*servicediscovery.ServiceSummary, 0)
		if id, present := sdNamespaces[name]; present {
			fmt.Printf("Discovering scraping targets in the namespace '%s'\n", name)
			filterType := aws.String("NAMESPACE_ID")
			filterCondition := aws.String("EQ")
			filterValues := []*string{&id}
			serviceFilter := servicediscovery.ServiceFilter{
				Name:      filterType,
				Values:    filterValues,
				Condition: filterCondition}
			listServiceOutput, err := c.service.ListServices(&servicediscovery.ListServicesInput{Filters: []*servicediscovery.ServiceFilter{&serviceFilter}})
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, serviceSummary := range listServiceOutput.Services {
				sdServices = append(sdServices, serviceSummary)
				sdServicesCount++
			}
			sdServicesMap[name] = sdServices
		}
	}
	fmt.Printf("No.of services discovered for scraping = %d\n", sdServicesCount)
	return sdServicesMap, nil
}

// Retrieve the list of tags associated with each ServiceDiscovery service.
// Tags are used to specify the URL path and port for endpoint where metrics are scraped from
// We are resorting to using tags because ServiceDiscovery API does not yet support adding custom attributes
func (c *CloudMapClient) getServiceTags(summary *servicediscovery.ServiceSummary) map[string]*string {
	tags := make(map[string]*string)
	getListTagsForResourceOutput, _ := c.service.ListTagsForResource(&servicediscovery.ListTagsForResourceInput{ResourceARN: summary.Arn})
	for _, serviceTag := range getListTagsForResourceOutput.Tags {
		tags[*serviceTag.Key] = serviceTag.Value
	}
	return tags
}

// Retrieve the list of ServiceDiscovery instances associated with each ServiceDiscovery service
// For each ServiceDiscovery instance, retrieve the default ECS attributes
func (c *CloudMapClient) getInstances(serviceSummary *servicediscovery.ServiceSummary) ([]*ServiceDiscoveryInstance, error) {
	sdInstaces := make([]*ServiceDiscoveryInstance, 0)
	getListInstancesOutput, err := c.service.ListInstances(&servicediscovery.ListInstancesInput{ServiceId: serviceSummary.Id})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, instanceSummary := range getListInstancesOutput.Instances {
		sdInstance := ServiceDiscoveryInstance{service: serviceSummary.Name, instanceId: instanceSummary.Id, attributes: instanceSummary.Attributes}
		sdInstaces = append(sdInstaces, &sdInstance)
	}
	fmt.Printf("No.of instances discovered for scraping in service '%s' = %d\n", *serviceSummary.Name, len(sdInstaces))
	return sdInstaces, nil
}

// Construct Prometheus scrape configuration for each ServiceDiscovery instance based on its attributes and the associated ServiceDiscovery service tags
func (c *CloudMapClient) getInstanceScrapeConfigurationApplication(sdInstance *ServiceDiscoveryInstance, serviceTags map[string]*string, sdNamespace *string) (*InstanceScrapeConfig, error) {
	metricsScheme, present := sdInstance.attributes[MetricsSchemeTag]
	if !present {
		metricsScheme = aws.String("http")
	}

	// Path for application metrics endpoint is expected as a resource tag with the key 'METRICS_PATH'
	metricsPath, present := serviceTags[MetricsPathTag]
	if !present {
		return nil, nil
	}

	// This is relevant for ECS tasks using bridge networking mode that are using host->container port mapping.
	// Port number of the resource is available, by default, as an attribute with the key 'AWS_INSTANCE_PORT'
	defaultPort, present := sdInstance.attributes[PortNumberAttribute]
	if !present {
		defaultPort = aws.String("80")
	}

	// Application metrics port is expected as a resource tag with the key 'METRICS_PORT'
	metricsPort, present := serviceTags[MetricsPortTag]
	if !present {
		metricsPort = defaultPort
	}

	return c.getInstanceScrapeConfiguration(sdInstance, metricsScheme, metricsPort, metricsPath, sdNamespace)
}

// This is relevant when the application is deployed along with a side-car container which exposes Docker stats as Prometheus metrics
// The Docker stats are available at the Task metadata endpoint ${ECS_CONTAINER_METADATA_URI_V4}/stats
// https://github.com/prometheus-community/ecs_exporter provides an implementation of a side-car that exposes Docker stats as Prometheus metrics
func (c *CloudMapClient) getInstanceScrapeConfigurationInfrastructure(sdInstance *ServiceDiscoveryInstance, serviceTags map[string]*string, sdNamespace *string) (*InstanceScrapeConfig, error) {
	// Path for infrastructure metrics endpoint is expected as a resource tag with the key 'ECS_METRICS_PATH'
	metricsPath, present := serviceTags[EcsMetricsPathTag]
	if !present {
		return nil, nil
	}

	// Infrastructure Metrics port is expected as a resource tag with the key 'ECS_METRICS_PORT'
	metricsPort, present := serviceTags[EcsMetricsPortTag]
	if !present {
		return nil, nil
	}

	// FIXME
	metricsScheme := aws.String("http")
	return c.getInstanceScrapeConfiguration(sdInstance, metricsScheme, metricsPort, metricsPath, sdNamespace)
}

func (c *CloudMapClient) getInstanceScrapeConfiguration(sdInstance *ServiceDiscoveryInstance, metricsScheme *string, metricsPort *string, metricsPath *string, sdNamespace *string) (*InstanceScrapeConfig, error) {
	labels := make(map[string]string)
	targets := make([]string, 0)

	// IP address of the resource is available, by default, as an attribute with the key 'AWS_INSTANCE_IPV4'
	address, present := sdInstance.attributes[IpAddressAttribute]
	if !present {
		return nil, errors.New(fmt.Sprintf("Cannot find IP address for instance in service %v", sdInstance.service))
	}
	targets = append(targets, fmt.Sprintf("%s:%s", *address, *metricsPort))

	//
	// ECS Task instances registered in Cloud Map are assigned the following default attributes
	// ECS_CLUSTER_NAME, ECS_SERVICE_NAME, ECS_TASK_DEFINITION_FAMILY
	// Add these attributes as labels to be attached to the Prometheus metric
	//
	cluster, present := sdInstance.attributes[ClusterNameAttribute]
	if present {
		labels["cluster"] = *cluster
	}
	service, present := sdInstance.attributes[ServiceNameAttribute]
	if present {
		labels["service"] = *service
	}
	taskdefinition, present := sdInstance.attributes[TaskDefinitionAttribute]
	if present {
		labels["taskdefinition"] = *taskdefinition
	}
	labels["namespace"] = *sdNamespace
	labels["taskid"] = *sdInstance.instanceId
	labels["instance"] = *address
	labels["__metrics_path__"] = *metricsPath
	labels["scheme"] = *metricsScheme

	return &InstanceScrapeConfig{Targets: targets, Labels: labels}, nil
}
