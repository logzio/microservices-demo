#!/bin/bash

echo "Scaling down load"

kubectl apply -f ~/kubernetes/manifests-loadtest/loadtest-dep.yaml
