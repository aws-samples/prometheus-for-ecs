##!/bin/bash
#
# Delete AMP workspace
#
aws amp delete-workspace --workspace-id $WORKSPACE_ID
#
# Delete SSM parameters
#
aws ssm delete-parameter --name otel-collector-config 
aws ssm delete-parameter --name ECS-Namespaces


