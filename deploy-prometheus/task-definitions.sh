##!/bin/bash

#
# Task Definitons
#
sed -e s/ACCOUNT/$ACCOUNT_ID/g \
-e s/REGION/$AWS_REGION/g \
< webappTaskDefinition.json.template \
> webappTaskDefinition.json
WEBAPP_TASK_DEFINITION=$(aws ecs register-task-definition \
--cli-input-json file://webappTaskDefinition.json \
--region $AWS_REGION \
--query "taskDefinition.taskDefinitionArn" --output text)

sed -e s/ACCOUNT/$ACCOUNT_ID/g \
-e s/REGION/$AWS_REGION/g \
< prometheusTaskDefinition.json.template \
> prometheusTaskDefinition.json
PROMETHEUS_TASK_DEFINITION=$(aws ecs register-task-definition \
--cli-input-json file://prometheusTaskDefinition.json \
--region $AWS_REGION \
--query "taskDefinition.taskDefinitionArn" --output text)

sed -e s/ACCOUNT/$ACCOUNT_ID/g \
-e s/REGION/$AWS_REGION/g \
< nodeExporterTaskDefinition.json.template \
> nodeExporterTaskDefinition.json
NODEEXPORTER_TASK_DEFINITION=$(aws ecs register-task-definition \
--cli-input-json file://nodeExporterTaskDefinition.json \
--region $AWS_REGION \
--query "taskDefinition.taskDefinitionArn" --output text)

export WEBAPP_TASK_DEFINITION
export PROMETHEUS_TASK_DEFINITION
export NODEEXPORTER_TASK_DEFINITION