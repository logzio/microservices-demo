#!/bin/bash

echo "Scaling down load"

URL=http://$(kubectl get services -n sock-shop front-end -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

kubectl patch configmap/loadtest-configmap -n loadtest --type merge -p '{"data":{"TARGET_HOST":"'"$URL"'","CLIENTS":"1","HATCH_RATE":"1","NUM_REQUEST":"8"}}'

kubectl rollout restart deployment load-test -n loadtest
