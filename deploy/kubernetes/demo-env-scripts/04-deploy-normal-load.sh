#!/bin/bash

echo "Scaling down load"

kubectl apply -f ../manifests-loadtest/loadtest-dep.yaml
