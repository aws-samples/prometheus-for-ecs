##!/bin/bash

#
# Delete CloudMap service registries and namespace
#
aws servicediscovery delete-service --id $CLOUDMAP_WEBAPP_SERVICE_ID
aws servicediscovery delete-service --id $CLOUDMAP_NODE_EXPORTER_SERVICE_ID
aws servicediscovery delete-namespace --id $CLOUDMAP_NAMESPACE_ID
