##!/bin/bash

#
# Create a trust policy for ECS task and task execution roles
#
cat <<EOF > TrustPolicy.json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

#
# Create a permission policy for the Task Execution Role
#
cat <<EOF > TaskExecutionPermissionPolicy.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ssm:GetParameter",
                "ssm:GetParameters"
            ],
            "Resource": "*"
        }
    ]
}
EOF


#
# Create a permission policy for for the Task role associated with the ADOT task
#
cat <<EOF > AdotTaskPermissionPolicy.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "aps:RemoteWrite",
                "aps:GetSeries",
                "aps:GetLabels",
                "aps:GetMetricMetadata"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "ssm:GetParameter",
                "ssm:GetParameters"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "servicediscovery:*"
            ],
            "Resource": "*"
        }
    ]
}
EOF

CLOUDWATCH_LOGS_POLICY_ARN=arn:aws:iam::aws:policy/CloudWatchLogsFullAccess
ECS_TASK_EXECUTION_POLICY_ARN=arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy


ECS_TASK_EXECUTION_ROLE="ECS-Task-Execution-Role"
ECS_TASK_EXECUTION_ROLE_ARN=$(aws iam create-role \
--role-name $ECS_TASK_EXECUTION_ROLE \
--assume-role-policy-document file://TrustPolicy.json \
--query "Role.Arn" --output text)

ECS_SSM_TASK_EXECUTION_POLICY="ECSSSMTaskExecutionPolicy"
ECS_SSM_TASK_EXECUTION_POLICY_ARN=$(aws iam create-policy --policy-name $ECS_SSM_TASK_EXECUTION_POLICY \
  --policy-document file://TaskExecutionPermissionPolicy.json \
  --query 'Policy.Arn' --output text)  

aws iam attach-role-policy --role-name $ECS_TASK_EXECUTION_ROLE --policy-arn $CLOUDWATCH_LOGS_POLICY_ARN
aws iam attach-role-policy --role-name $ECS_TASK_EXECUTION_ROLE --policy-arn $ECS_TASK_EXECUTION_POLICY_ARN
aws iam attach-role-policy --role-name $ECS_TASK_EXECUTION_ROLE --policy-arn $ECS_SSM_TASK_EXECUTION_POLICY_ARN

ECS_GENERIC_TASK_ROLE="ECS-Generic-Task-Role"
ECS_GENERIC_TASK_ROLE_ARN=$(aws iam create-role \
--role-name $ECS_GENERIC_TASK_ROLE \
--assume-role-policy-document file://TrustPolicy.json \
--query "Role.Arn" --output text)
aws iam attach-role-policy --role-name $ECS_GENERIC_TASK_ROLE --policy-arn $CLOUDWATCH_LOGS_POLICY_ARN

ECS_ADOT_TASK_ROLE="ECS-ADOT-Task-Role"
ECS_ADOT_TASK_ROLE_ARN=$(aws iam create-role \
--role-name $ECS_ADOT_TASK_ROLE \
--assume-role-policy-document file://TrustPolicy.json \
--query "Role.Arn" --output text)

ECS_ADOT_TASK_POLICY="ECSAdotTaskPolicy"
ECS_ADOT_TASK_POLICY_ARN=$(aws iam create-policy --policy-name $ECS_ADOT_TASK_POLICY \
  --policy-document file://AdotTaskPermissionPolicy.json \
  --query 'Policy.Arn' --output text)

aws iam attach-role-policy --role-name $ECS_ADOT_TASK_ROLE --policy-arn $CLOUDWATCH_LOGS_POLICY_ARN
aws iam attach-role-policy --role-name $ECS_ADOT_TASK_ROLE --policy-arn $ECS_ADOT_TASK_POLICY_ARN  

export ECS_GENERIC_TASK_ROLE
export ECS_TASK_EXECUTION_ROLE
export ECS_ADOT_TASK_ROLE

export CLOUDWATCH_LOGS_POLICY_ARN
export ECS_TASK_EXECUTION_POLICY_ARN
export ECS_ADOT_TASK_POLICY_ARN
