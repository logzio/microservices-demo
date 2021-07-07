# Deploying sock-shop to a Kubernetes cluster in AWS

## Option 1 - Create a new Kubernetes cluster via the EKS CLI

### Prerequisites

1. `kubectl` and `eksctl` set up on your system - see the _Prerequisites_ section in the [Getting started with Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html) guide
2. An SSH keypair set up for the AWS reqion you wish to create the EKS stack in - see [Creating or importing a keypair](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html#prepare-key-pair)

### Create the cluster

In a terminal on your local machine, run the following command:

```bash
eksctl create cluster --name <cluster-name> --region <aws-region> --with-oidc --ssh-access --ssh-public-key <ssh-key-name> --managed --zones <zone-list> --node-type t3.medium
```

- `<cluster-name>` is the name you wish to identify your cluster with
- `<aws-region>` is the name of a valid [AWS Region](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-regions)
- `<ssh-key-name>` is the name of a valid SSH key in the `<aws-region>`
- `<zone-list>` is a comma-separated list of two [Availability Zones](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#availability-zones-describe) in the `<aws-region>` e.g. `us-east-1a,us-east-1b`
  
For example:
```bash 
eksctl create cluster --name sf-demo --region us-east-1 --with-oidc --ssh-access --ssh-public-key simon.fisher --managed --zones us-east-1a,us-east-1b --node-type t3.medium
```

Once the cluster is created, verify the connection by running `kubectl get nodes`

## Option 2 - Create a new Kubernetes cluster in AWS via Console

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

### Connect to and test the running cluster:

1. On your local machine switch to your home directory `cd ~`
2. Export the path to the private key file for the keypair you used in the stack creation:
   `SSH_KEY="<PATH_TO_PRIVATE_AWS_KEY.PEM>"`
3. Look up the `BastionHostPublicIp` and `MasterPrivateIp` values in the Cloudformation Outputs page and then run:
   `scp -i $SSH_KEY -o ProxyCommand="ssh -i \"${SSH_KEY}\" ubuntu@<BastionHostPublicIp> nc %h %p" ubuntu@<MasterPrivateIp>:~/kubeconfig ./kubeconfig`
4. `export KUBECONFIG=$(pwd)/kubeconfig`
5. Verify the connection by running `kubectl get nodes`

---

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

Otionally update your region if you are using a different region than the default in the exporters section in `deploy/kubernetes/manifests-tracing/otel-collector.yaml` 

```
cd microservices-demo

kubectl --namespace=monitoring create secret generic logzio-tracing-secret \
  --from-literal=logzio-tracing-shipping-token='<<TRACING-SHIPPING-TOKEN>>' 

kubectl apply -f deploy/kubernetes/manifests-tracing
```
---

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