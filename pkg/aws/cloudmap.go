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
	ScrapeConfigParmeter    = "ECS-Scrape-Configuration"
	IpAddressAttribute      = "AWS_INSTANCE_IPV4"
	ClusterNameAttribute    = "ECS_CLUSTER_NAME"
	ServiceNameAttribute    = "ECS_SERVICE_NAME"
	TaskDefinitionAttribute = "ECS_TASK_DEFINITION_FAMILY"
	MetricsPortTag          = "METRICS_PORT"
	MetricsPathTag          = "METRICS_PATH"
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

//
// Retrieve a JSON object that provides a list of ECS targets to be scraped for Prometheus metrics
//
func GetPrometheusScrapeConfig(selectedNamespaces []string) *string {
	client := &CloudMapClient{service: servicediscovery.New(sharedSession)}

	sdNamespaces, _ := client.getNamespaces()
	sdServices, _ := client.getServices(selectedNamespaces, sdNamespaces)
	scrapeConfigurations := make([]*InstanceScrapeConfig, 0)
	for _, service := range sdServices {
		serviceTags := client.getServiceTags(service)
		sdInstances, _ := client.getInstances(service)
		for _, instance := range sdInstances {
			scrapeConfig, _ := client.getInstanceScrapeConfiguration(instance, serviceTags)
			scrapeConfigurations = append(scrapeConfigurations, scrapeConfig)
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

//
// Get a list of all ServiceDiscovery namespaces and their respective IDs available under Cloud Map
//
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

//
// Cycle through each ServiceDiscovery namespace and find the list of ServiceDiscovery services
//
func (c *CloudMapClient) getServices(selectedNamespaces []string, sdNamespaces map[string]string) ([]*servicediscovery.ServiceSummary, error) {
	sdServices := make([]*servicediscovery.ServiceSummary, 0)
	for _, name := range selectedNamespaces {
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
			}
		}
	}
	fmt.Printf("No.of services discovered for scraping = %d\n", len(sdServices))
	return sdServices, nil
}

//
// Retrieve the list of tags associated with each ServiceDiscovery service.
// Tags are used to specify the URL path and port for endpoint where metrics are scraped from
// We are resorting to using tags because ServiceDiscovery API does not yet support adding custom attributes
//
func (c *CloudMapClient) getServiceTags(summary *servicediscovery.ServiceSummary) map[string]*string {
	tags := make(map[string]*string)
	getListTagsForResourceOutput, _ := c.service.ListTagsForResource(&servicediscovery.ListTagsForResourceInput{ResourceARN: summary.Arn})
	for _, serviceTag := range getListTagsForResourceOutput.Tags {
		tags[*serviceTag.Key] = serviceTag.Value
	}
	return tags
}

//
// Retrieve the list of ServiceDiscovery instances associated with each ServiceDiscovery service
// For each ServiceDiscovery instance, retrieve the default ECS attributes
//
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

//
// Construct Prometheus scrape configuration for each ServiceDiscovery instance based on its attributes and the associated ServiceDiscovery service tags
//
func (c *CloudMapClient) getInstanceScrapeConfiguration(sdInstance *ServiceDiscoveryInstance, serviceTags map[string]*string) (*InstanceScrapeConfig, error) {
	labels := make(map[string]string)
	targets := make([]string, 0)

	// Metrics port is expected as a resource tag with the key 'METRICS_PORT'
	defaultPort := aws.String("80")
	metricsPort, present := serviceTags[MetricsPortTag]
	if !present {
		metricsPort = defaultPort
	}

	// IP address of the resource is available, by default, as an attribute with the key 'AWS_INSTANCE_IPV4'
	address, present := sdInstance.attributes[IpAddressAttribute]
	if !present {
		return nil, errors.New(fmt.Sprintf("Cannot find IP address for instance in service %v", sdInstance.service))
	}
	targets = append(targets, fmt.Sprintf("%s:%s", *address, *metricsPort))

	// Path for metrics endpoint is expected as a resource tag with the key '__metrics_path__'
	defaultPath := aws.String("/metrics")
	metricsPath, present := serviceTags[MetricsPathTag]
	if !present {
		metricsPath = defaultPath
	}
	labels["__metrics_path__"] = *metricsPath

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

	return &InstanceScrapeConfig{Targets: targets, Labels: labels}, nil
}
