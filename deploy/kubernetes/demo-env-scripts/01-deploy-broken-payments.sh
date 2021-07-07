#!/bin/bash

while getopts a: flag
do
    case "${flag}" in
        a) api_token=${OPTARG};;
    esac
done

echo "Deploying version 0.4.4 of payments service"

kubectl set image deployment/payment payment=850759972217.dkr.ecr.us-east-1.amazonaws.com/payment:0.4.4 --record -n sock-shop

echo "Notifying Logz.io of deployment"

curl --header "Content-Type: application/json" --header "X-API-TOKEN: $api_token" \
  --request POST \
  --data '{"markers": [{"description": "Deploying version 0.4.4 of Payments service","title": "Deployed version 0.4.4 of Payments service","tag": "DEPLOYMENT","metadata": {"version": "0.4.4","region": "us-east-1","deployedBy": "Tomer"}}]}' \
  https://api.logz.io/v2/markers/create-markers