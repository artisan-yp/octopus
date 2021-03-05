#!/bin/bash

kubectl create --save-config configmap myapp --from-file=config.toml=./myapp_config.toml -o yaml --dry-run=client -n test | kubectl apply -f -
