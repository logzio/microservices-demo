#!/bin/bash

while getopts l:m:t: flag
do
    case "${flag}" in
        l) logs_token=${OPTARG};;
        m) metrics_token=${OPTARG};;
        t) tracing_token=${OPTARG};;
    esac
done

echo "Deploying fluentd logging"

kubectl create namespace monitoring

kubectl create secret generic logzio-logs-secret \
  --from-literal=logzio-log-shipping-token=$logs_token \
  --from-literal=logzio-log-listener='https://listener.logz.io:8071' \
  -n monitoring

kubectl apply -f https://raw.githubusercontent.com/logzio/logzio-k8s/master/logzio-daemonset-rbac.yaml -f https://raw.githubusercontent.com/logzio/logzio-k8s/master/configmap.yaml

echo "Deploying Prometheus monitoring"

gsed -i "s/METRICS_SHIPPING_TOKEN/$metrics_token/g" ../manifests-monitoring/prometheus-configmap.yaml

kubectl apply -f ../manifests-monitoring

echo "Deploying OpenTelemetry tracing"

kubectl --namespace=monitoring create secret generic logzio-tracing-secret \
  --from-literal=logzio-tracing-shipping-token=$tracing_token 

kubectl apply -f ../manifests-tracing

echo "Deploying Sock Shop"

kubectl apply -f ../manifests

echo "Starting normal traffic"

kubectl apply -f ../manifests-loadtest/loadtest-dep.yaml