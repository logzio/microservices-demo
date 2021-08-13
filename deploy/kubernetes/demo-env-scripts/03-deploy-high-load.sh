#!/bin/bash

echo "Scaling up load"

URL=http://$(kubectl get services -n sock-shop front-end -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

kubectl patch configmap/loadtest-configmap -n loadtest --type merge -p '{"data":{"TARGET_HOST":"'"$URL"'","CLIENTS":"4","HATCH_RATE":"4","NUM_REQUEST":"80"}}'

kubectl rollout restart deployment load-test -n loadtest
