#!/bin/bash

echo "Scaling up load"

kubectl apply -f ~/kubernetes/manifests-loadtest/loadtest-dep-high-traffic.yaml
