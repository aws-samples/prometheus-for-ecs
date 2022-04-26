##!/bin/bash

#
# Create a Service Discovery namespace
#
SERVICE_DISCOVERY_NAMESPACE=ecs-services
OPERATION_ID=$(aws servicediscovery create-private-dns-namespace \
--vpc $VPC_ID \
--name $SERVICE_DISCOVERY_NAMESPACE \
--query "OperationId" --output text)

operationStatus() {
  aws servicediscovery get-operation --operation-id $OPERATION_ID --query "Operation.Status" --output text
}

until [ $(operationStatus) != "PENDING" ]; do
  echo "Namespace $SERVICE_DISCOVERY_NAMESPACE is creating ..."
  sleep 10s
  if [ $(operationStatus) = "SUCCESS" ]; then
    echo "Namespace $SERVICE_DISCOVERY_NAMESPACE created"
    break
  fi
done

CLOUDMAP_NAMESPACE_ID=$(aws servicediscovery get-operation \
--operation-id $OPERATION_ID \
--query "Operation.Targets.NAMESPACE" --output text)

#
# Create a Service Discovery service in the above namespace
# When create a Service Discovery service with either private or public DNS, there are different options available for DNS record type.
# When doing a DNS query on the service name:
#   1. "A" records return a set of IP addresses that correspond to your tasks. 
#   2. "SRV" records return a set of IP addresses and ports per task.
#
METRICS_PATH=/metrics
METRICS_PORT=3000
ECS_METRICS_PATH=/metrics
ECS_METRICS_PORT=9779
SERVICE_REGISTRY_NAME="webapp-svc"
SERVICE_REGISTRY_DESCRIPTION="Service registry for Webapp ECS service"
CLOUDMAP_WEBAPP_SERVICE_ID=$(aws servicediscovery create-service \
--name $SERVICE_REGISTRY_NAME \
--description "$SERVICE_REGISTRY_DESCRIPTION" \
--namespace-id $CLOUDMAP_NAMESPACE_ID \
--dns-config "NamespaceId=$CLOUDMAP_NAMESPACE_ID,RoutingPolicy=WEIGHTED,DnsRecords=[{Type=A,TTL=10}]" \
--region $AWS_REGION \
--tags Key=METRICS_PATH,Value=$METRICS_PATH Key=METRICS_PORT,Value=$METRICS_PORT Key=ECS_METRICS_PATH,Value=$ECS_METRICS_PATH Key=ECS_METRICS_PORT,Value=$ECS_METRICS_PORT \
--query "Service.Id" --output text)
CLOUDMAP_WEBAPP_SERVICE_ARN=$(aws servicediscovery get-service \
--id $CLOUDMAP_WEBAPP_SERVICE_ID \
--query "Service.Arn" --output text)
echo "Service registry $SERVICE_REGISTRY_NAME created"

SERVICE_REGISTRY_NAME="adot-collector-svc"
SERVICE_REGISTRY_DESCRIPTION="Service registry for ADOT Collector ECS service"
CLOUDMAP_ADOT_COLLECTOR_SERVICE_ID=$(aws servicediscovery create-service \
--name $SERVICE_REGISTRY_NAME \
--description "$SERVICE_REGISTRY_DESCRIPTION" \
--namespace-id $CLOUDMAP_NAMESPACE_ID \
--dns-config "NamespaceId=$CLOUDMAP_NAMESPACE_ID,RoutingPolicy=WEIGHTED,DnsRecords=[{Type=A,TTL=10}]" \
--region $AWS_REGION \
--query "Service.Id" --output text)
CLOUDMAP_ADOT_COLLECTOR_SERVICE_ARN=$(aws servicediscovery get-service \
--id $CLOUDMAP_ADOT_COLLECTOR_SERVICE_ID \
--query "Service.Arn" --output text)
echo "Service registry $SERVICE_REGISTRY_NAME created"

export CLOUDMAP_NAMESPACE_ID
export CLOUDMAP_WEBAPP_SERVICE_ARN
export CLOUDMAP_WEBAPP_SERVICE_ID
export CLOUDMAP_ADOT_COLLECTOR_SERVICE_ARN
export CLOUDMAP_ADOT_COLLECTOR_SERVICE_ID