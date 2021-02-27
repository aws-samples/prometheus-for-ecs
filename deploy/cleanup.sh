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
# Delete CloudMap service registries and namespace
#
aws servicediscovery delete-service --id $CLOUDMAP_WEBAPP_SERVICE_ID
aws servicediscovery delete-service --id $CLOUDMAP_NODE_EXPORTER_SERVICE_ID
aws servicediscovery delete-namespace --id $CLOUDMAP_NAMESPACE_ID

#
# Deregister task definitions
#
aws ecs deregister-task-definition --task-definition $WEBAPP_TASK_DEFINITION
aws ecs deregister-task-definition --task-definition $PROMETHEUS_TASK_DEFINITION
aws ecs deregister-task-definition --task-definition $NODEEXPORTER_TASK_DEFINITION


#
# Delete IAM roles and policies
#
aws iam detach-role-policy --role-name $ECS_GENERIC_TASK_ROLE --policy-arn $CLOUDWATCH_LOGS_POLICY_ARN
aws iam detach-role-policy --role-name $ECS_TASK_EXECUTION_ROLE --policy-arn $CLOUDWATCH_LOGS_POLICY_ARN
aws iam detach-role-policy --role-name $ECS_TASK_EXECUTION_ROLE --policy-arn $ECS_TASK_EXECUTION_POLICY_ARN
aws iam detach-role-policy --role-name $ECS_PROMETHEUS_TASK_ROLE --policy-arn $ECS_PROMETHEUS_TASK_POLICY_ARN

aws iam delete-policy --policy-arn $ECS_PROMETHEUS_TASK_POLICY_ARN

aws iam delete-role --role-name $ECS_GENERIC_TASK_ROLE
aws iam delete-role --role-name $ECS_TASK_EXECUTION_ROLE
aws iam delete-role --role-name $ECS_PROMETHEUS_TASK_ROLE

#
# Delete the CloudFormation stack
#
