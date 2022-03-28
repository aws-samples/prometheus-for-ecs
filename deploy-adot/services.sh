##!/bin/bash

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
--enable-execute-command \
--service-registries "registryArn=$CLOUDMAP_SERVICE_ARN" \
--network-configuration "awsvpcConfiguration={subnets=$PRIVATE_SUBNET_IDS,securityGroups=[$SECURITY_GROUP_ID],assignPublicIp=DISABLED}" \
--scheduling-strategy REPLICA \
--launch-type EC2

#
# Create the ADOT Service
#
SERVICE_NAME=ADOTService
TASK_DEFINITION=$ADOT_TASK_DEFINITION
aws ecs create-service --service-name $SERVICE_NAME \
--cluster $CLUSTER_NAME \
--task-definition $TASK_DEFINITION \
--desired-count 1 \
--enable-execute-command \
--network-configuration "awsvpcConfiguration={subnets=$PRIVATE_SUBNET_IDS,securityGroups=[$SECURITY_GROUP_ID],assignPublicIp=DISABLED}" \
--scheduling-strategy REPLICA \
--launch-type EC2