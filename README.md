## Custom metrics collection using Prometheus on Amazon ECS

This Git repository contains software artifacts to deploy Prometheus server and Prometheus Node Exporter to an Amazon ECS cluster. The Golang code pertains to that of a side-car container that is deployed alongside the Prometheus server in an ECS task and it enables discovery of scraping targets in the ECS cluster.

<img class="wp-image-1960 size-full" src="images/Depoloyment-Architecture.png" alt="Deployment architecture" width="854" height="527" />

### Solution overview

At a high level, we will be following the steps outlined below for this solution:

<ul>
  <li>
    Setup AWS Cloud Map for service discovery 
  </li>
  <li>
    Deploy application services to an Amazon ECS and register them with AWS Cloud Map
  </li>
  <li>
    Deploy Prometheus server to Amazon ECS, configure service discovery and send metrics data to Amazon Managed Service for Prometheus
  </li>
  <li>
    Visualize metrics data using Amazon Managed Service for Grafana
  </li>  
</ul>

### Deploy

Requires an ECS cluster. For deploying the Prometheus Node Exporter, a cluster with EC2 instances is required. All deployment artifacts are under the [deploy](https://github.com/aws-samples/prometheus-for-ecs/tree/main/deploy) directory. The deployment comprises the following components:
- An ECS task comprising the Prometheus server, AWS Sig4 proxy and the service discovery application containers

- A mock web application that is instrumented with Prometheus client library and exposes an HTTP endpoint */work*. The application has an internal load generator that sends client requests to the HTTP endpoint. The service exposes a [Counter](https://prometheus.io/docs/concepts/metric_types/#counter) named *http_requests_total* and a [Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram) named *request_durtaion_milliseconds*
 
- Prometheus Node Exporter to monitor system metrics from every container instance in the cluster. This service is deployed using [*host* networking mode](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#network_mode) and with the daemon scheduling strategy. Note that we can’t deploy the Node Exporter on AWS Fargate because it does not support the daemon scheduling strategy.


The deploment scripts assume that the underlying ECS cluster was created using the [ecs-cluster.yaml](https://github.com/aws-samples/prometheus-for-ecs/blob/main/deploy/ecs-cluster.yaml) CloudFormation template. Create the cluster with the following command:
``` 
VPC_STACK_NAME=ecs-stack 
VPC_TEMPLATE=ecs-cluster.yaml
aws cloudformation deploy --stack-name $VPC_STACK_NAME --template-file $VPC_TEMPLATE --capabilities CAPABILITY_IAM 
```
    
First, export a set of environment variables that are required by scripts used in subsequent steps. Modify the **ACCOUNT_ID** and **AWS_REGION** variables in the script before running the command below.
```
source env.sh
```

Create the ECS task and task execution roles and the relevenat IAM policies.
```
source iam.sh
```

Create a service discovery namespace and service registries under AWS Cloud Map
```
source cloudmap.sh
```

Create a workspace under Amazon Managed Service for Prometheus (AMP) for ingesting Prometheus metrics scraped from ECS services.
```
source amp.sh
```
The above command generates the initial configuration file *prometheus.yaml* for the Prometheus server, with the AMP worksapce as the remote write destination. 
Create two parameters in the AWS SSM Parameter as follows:
- parameter named **ECS-Prometheus-Configuration** and of type *String* using the contents of the prometheus.yaml file
- parameter named **ECS-ServiceDiscovery-Namespaces** and of type *String* with the value set to **ecs-services**

Next, register task definitions with ECS
```
source task-definitions.sh
```

Launch the ECS services using the task definitions created above. 
```
source services.sh
```

Use Amazon Managed Service for Grafana to query and visualize the metrics ingested into AMP. You may use the following PromQL queries to visualize the metrics collected from the web application and Prometheus Node Exporter
- HTTP request rate: *sum(rate(http_requests_total[5m]))*
- Average response latency: *sum(rate(request_durtaion_milliseconds_sum[5m])) / sum(rate(request_durtaion_milliseconds_count[5m]))*
- Average CPU usage:  *100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[1m])) * 100)*

### Cleanup

When you are done, cleanup the resources you created above with the follwing set of commands.
```
source cleanup-ecs.sh
source cleanup-cloudmap.sh
source cleanup-iam.sh
aws cloudformation delete-stack --stack-name $VPC_STACK_NAME
```

## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This library is licensed under the MIT-0 License. See the LICENSE file.

