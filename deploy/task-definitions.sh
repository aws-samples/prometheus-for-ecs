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
< adotTaskDefinition.json.template \
> adotTaskDefinition.json
ADOT_TASK_DEFINITION=$(aws ecs register-task-definition \
--cli-input-json file://adotTaskDefinition.json \
--region $AWS_REGION \
--query "taskDefinition.taskDefinitionArn" --output text)

export WEBAPP_TASK_DEFINITION
export ADOT_TASK_DEFINITION