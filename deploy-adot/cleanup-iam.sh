##!/bin/bash
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

