#!/bin/bash

kubectl apply -f mysql_secret.yaml

kubectl apply -f myapp_deployment.yaml

