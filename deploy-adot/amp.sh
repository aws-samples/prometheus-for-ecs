##!/bin/bash
WORKSPACE_ID=$(aws amp create-workspace --alias prometheus-for-ecs --query "workspaceId" --output text)

sed -e s/WORKSPACE/$WORKSPACE_ID/g \
-e s/REGION/$AWS_REGION/g \
< otel-collector-config.yaml.template \
> otel-collector-config.yaml
