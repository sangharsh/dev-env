# Deploy app with istio routing

Deploy app and setup istio to route request to app version based on header

### Prerequisites

1. Kubernetes cluster e.g. [Minikube](https://minikube.sigs.k8s.io/docs/start/) (below example uses it)
1. Tool to build container image e.g. [Docker](https://www.docker.com/)
1. Service mesh like [Istio](https://istio.io/latest/docs/setup/getting-started/#download)

## Setup instructions

### Kubernetes cluster

```
minikube start --memory=4096 -p devenv
// Run in a separate window
// Assigns an external IP for LoadBalancer services
minikube tunnel -p devenv
```

### Service mesh

```
istioctl install --skip-confirmation
kubectl label namespace default istio-injection=enabled
```

### Container image

```
eval $(minikube docker-env -p devenv)
docker build -t hello:latest -f hello/Dockerfile hello/
```

### Deploy

```
kubectl apply -f istio/baseline.yaml
// Access from within a pod
kubectl exec "$(kubectl get pod -l app=hello-2 -o jsonpath='{.items[0].metadata.name}')" -c hello-2 -- wget -q -O- hello-1:8080/hello | jq
```

### Access the app

Get Gateway URL

```
source istio/set_gateway_url.sh
echo $GATEWAY_URL
```

Send a request to apps via ingress gateway
```
curl -sS ${GATEWAY_URL}/hello | jq
```

### Logs

Logs from app container

```
kubectl logs -f "$(kubectl get pod -l app=hello-2,version=baseline -o jsonpath='{.items[0].metadata.name}')" -c hello-2
```

Access logs from envoy sidecar

```
kubectl apply -f istio/telemetry.yaml
kubectl logs -f "$(kubectl get pod -l app=hello-2,version=baseline -o jsonpath='{.items[0].metadata.name}')" -c istio-proxy
```

## Clean up

```
minikube delete -p devenv
```
