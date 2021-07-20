#!/bin/bash

echo "Scaling up load"

URL=http://$(kubectl get services -n sock-shop front-end -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

kubectl patch configmap/loadtest-configmap -n loadtest --type merge -p '{"data":{"TARGET_HOST":"'"$URL"'","CLIENTS":"15","HATCH_RATE":"15","NUM_REQUEST":"20"}}'

kubectl rollout restart deployment load-test -n loadtest
