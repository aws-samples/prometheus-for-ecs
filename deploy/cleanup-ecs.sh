##!/bin/bash

#
# Delete the ECS services
#
SERVICE_NAME=WebAppService
aws ecs update-service --cluster $CLUSTER_NAME --service $SERVICE_NAME --desired-count 0
aws ecs delete-service --cluster $CLUSTER_NAME --service $SERVICE_NAME

SERVICE_NAME=PrometheusService
aws ecs update-service --cluster $CLUSTER_NAME --service $SERVICE_NAME --desired-count 0
aws ecs delete-service --cluster $CLUSTER_NAME --service $SERVICE_NAME

SERVICE_NAME=NodeExporterService
aws ecs delete-service --cluster $CLUSTER_NAME --service $SERVICE_NAME

#
# Deregister task definitions
#
aws ecs deregister-task-definition --task-definition $WEBAPP_TASK_DEFINITION
aws ecs deregister-task-definition --task-definition $PROMETHEUS_TASK_DEFINITION
aws ecs deregister-task-definition --task-definition $NODEEXPORTER_TASK_DEFINITION
