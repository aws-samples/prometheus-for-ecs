##!/bin/bash

#
# NodeExporter Service
#
SERVICE_NAME=NodeExporterService
TASK_DEFINITION=$NODEEXPORTER_TASK_DEFINITION
CLOUDMAP_SERVICE_ARN=$CLOUDMAP_NODE_EXPORTER_SERVICE_ARN
aws ecs create-service --service-name $SERVICE_NAME \
--cluster $CLUSTER_NAME \
--task-definition $TASK_DEFINITION \
--service-registries "containerName=prometheus-node-exporter,containerPort=9100,registryArn=$CLOUDMAP_SERVICE_ARN" \
--scheduling-strategy DAEMON \
--launch-type EC2

#
# WebApp Service
#
SERVICE_NAME=WebAppService
TASK_DEFINITION=$WEBAPP_TASK_DEFINITION
CLOUDMAP_SERVICE_ARN=$CLOUDMAP_WEBAPP_SERVICE_ARN
aws ecs create-service --service-name $SERVICE_NAME \
--cluster $CLUSTER_NAME \
--task-definition $TASK_DEFINITION \
--desired-count 2 \
--service-registries "registryArn=$CLOUDMAP_SERVICE_ARN" \
--network-configuration "awsvpcConfiguration={subnets=$PRIVATE_SUBNET_IDS,securityGroups=[$SECURITY_GROUP_ID],assignPublicIp=DISABLED}" \
--scheduling-strategy REPLICA \
--launch-type EC2

#
# Create the Prometheus Service
#
SERVICE_NAME=PrometheusService
TASK_DEFINITION=$PROMETHEUS_TASK_DEFINITION
aws ecs create-service --service-name $SERVICE_NAME \
--cluster $CLUSTER_NAME \
--task-definition $TASK_DEFINITION \
--desired-count 1 \
--network-configuration "awsvpcConfiguration={subnets=$PRIVATE_SUBNET_IDS,securityGroups=[$SECURITY_GROUP_ID],assignPublicIp=DISABLED}" \
--scheduling-strategy REPLICA \
--launch-type EC2