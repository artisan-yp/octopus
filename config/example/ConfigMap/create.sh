#!/bin/bash

# create configmap
kubectl create configmap myapp --from-file=config.toml=./myapp_config.toml -n test

# create deployment
kubectl apply -f myapp_deployment.yaml

