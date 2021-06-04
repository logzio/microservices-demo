#!/bin/bash

echo "Scaling up load"

kubectl apply -f ../manifests-loadtest/loadtest-dep-high-traffic.yaml
