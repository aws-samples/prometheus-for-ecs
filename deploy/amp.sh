##!/bin/bash
WORKSPACE_ID=$(aws amp create-workspace --alias prometheus-for-ecs --query "workspaceId" --output text)

sed -e s/WORKSPACE/$WORKSPACE_ID/g \
< prometheus.yaml.template \
> prometheus.yaml
