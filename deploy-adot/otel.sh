##!/bin/bash
sed -e s/WORKSPACE/$WORKSPACE_ID/g \
-e s/REGION/$AWS_REGION/g \
< otel-collector-config.yaml.template \
> otel-collector-config.yaml
