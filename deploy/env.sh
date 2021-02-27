
##!/bin/bash
export AWS_REGION=us-east-1
export ACCOUNT_ID=937351930975
export STACK_NAME=ecs-stack 

export CLUSTER_NAME=$(aws cloudformation describe-stacks --stack-name $STACK_NAME --query 'Stacks[0].Outputs[?OutputKey==`ClusterName`].OutputValue' --output text)
export VPC_ID=$(aws cloudformation describe-stacks --stack-name $STACK_NAME --query 'Stacks[0].Outputs[?OutputKey==`VpcId`].OutputValue' --output text)
export PUBLIC_SUBNET_IDS=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" "Name=tag-key,Values=ecs.io/role/elb"  --query "Subnets[].SubnetId" --output text)
export PRIVATE_SUBNET_IDS=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" "Name=tag-key,Values=ecs.io/role/internal-elb"  --query "Subnets[].SubnetId" --output json)
export SECURITY_GROUP_ID=$(aws cloudformation describe-stacks --stack-name $STACK_NAME --query 'Stacks[0].Outputs[?OutputKey==`ContainerSecurityGroup`].OutputValue' --output text)
