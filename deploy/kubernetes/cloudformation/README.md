# Deploying sock-shop to a Kubernetes cluster in AWS

## Create a new Kubernetes cluster in AWS

Use the `kubernetes-cluster-with-new-vpc.template.yaml` template in this directory to spin up a 3-node Kubernetes cluster in AWS.  This template is taken from the VMWare Kubernetes Quickstart (https://aws.amazon.com/quickstart/architecture/vmware-kubernetes/)

1. Sign in to the AWS console
2. Select the Cloudformation service
3. Click Create Stack > With new resources (standard)
4. Select Template is ready and then Upload a template file.  Choose the `kubernetes-cluster-with-new-vpc.template.yaml` contained in this repository.
5. Enter a stack name.  A good practice is to use your name or initials somewhere in the name so that a stack can be identified as yours.
6. Enter the following parameters (the rest can be left as their defaults):
   - **Availability Zone** - Choose any from the list
   - **Admin Ingress Location** - Use 0.0.0.0/0 to allow access from anywhere or a valid IP range to restrict access
   - **SSH Key** - The SSH keypair you use for access to EC2 instances.  You must have access to the private key on your local system.
7. Click _Next_ twice
8. Tick the boxes in the _Capabilities_ section and finally click the **Create Stack** button.
9. Once the creation is complete, note the contents of the **Outputs** tab - you will need some of these in the following sections.

## Connect to and test the running cluster:

1. On your local machine switch to your home directory `cd ~`
2. Export the path to the private key file for the keypair you used in the stack creation:
   `SSH_KEY="<PATH_TO_PRIVATE_AWS_KEY.PEM>"`
3. Look up the `BastionHostPublicIp` and `MasterPrivateIp` values in the Cloudformation Outputs page and then run:
   `scp -i $SSH_KEY -o ProxyCommand="ssh -i \"${SSH_KEY}\" ubuntu@<BastionHostPublicIp> nc %h %p" ubuntu@<MasterPrivateIp>:~/kubeconfig ./kubeconfig`
4. `export KUBECONFIG=$(pwd)/kubeconfig`
5. Verify the connection by running `kubectl get nodes`

## Deploy the Logz.io fluentd daemonset

Follow the instructions, using the __RBAC__ option: https://app.logz.io/#/dashboard/send-your-data/log-sources/kubernetes?type=default-config

Verify the monitoring stack is running (pods should be in the _Ready_ state)
```
kubectl get pods -n monitoring
```

## Deploy metrics shipping using Prometheus

Add your METRICS_SHIPPING_TOKEN to the remote write section in `deploy/kubernetes/manifests-monitoring/prometheus-configmap.yaml` and optionally update your listener URL address if you are using a different region than the default

**NOTE:**  We need to do this as prometheus.yml can't read environment variables.  Long term we need a better solution to avoid people sharing their shipping token in this repository

```
cd microservices-demo
kubectl apply -f deploy/kubernetes/manifests-monitoring
```
## Deploy tracing shipping using the OpenTelemetry Collector

Add your TRACING_SHIPPING_TOKEN to the exporters section in `deploy/kubernetes/manifests-tracing/otel-collector.yaml` and optionally update your region if you are using a different region than the default

**NOTE:**  We need to do this as otel-collector.yaml can't read environment variables.  Long term we need a better solution to avoid people sharing their shipping token in this repository

```
cd microservices-demo
kubectl apply -f deploy/kubernetes/manifests-tracing
```

## Deploy the sock-shop app stack

```
kubectl apply -f deploy/kubernetes/manifests
```

Wait approximately 5 minutess for DNS propagation to happen and then get the app's External IP address from the `LoadBalancer` service
```
kubectl get services -n sock-shop
```
You can now visit the sock-shop using this address in a web browser

## (Optional) Start and stop a load test

Start the load test:
```
kubectl apply -f deploy/kubernetes/manifests-loadtest
```
Stop the load test:
```
kubectl delete -f deploy/kubernetes/manifests-loadtest
```

## Clean down all of the services in the Kubernetes cluster

```
kubectl delete namespace sock-shop
kubectl delete namespace monitoring
```

## Delete the Cloudformation stack
1. Log in to the the AWS Console and find your stack in the Cloudformation service - ensure you choose the main (not nested) one
2. Click the __Delete__ button in the top right hand corner