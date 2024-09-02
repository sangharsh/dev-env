# Deploy app with istio routing

Deploy app and setup istio to route request to app version based on header

## Setup kubernetes

```
minikube start -p minikube
// Run in a separate window
// Assigns an external IP for LoadBalancer services
minikube tunnel -p minikube
```

## Install istio on k8s

```
istioctl install --skip-confirmation
kubectl label namespace default istio-injection=enabled
```

## Build image

```
eval $(minikube docker-env -p minikube)
docker build -t hello:latest -f hello/Dockerfile hello/
```

## Deploy

```
kubectl apply -f istio/deployments.yaml
// Access from within a pod
kubectl exec "$(kubectl get pod -l app=hello-2 -o jsonpath='{.items[0].metadata.name}')" -c hello-2 -- wget -q -O- hello-1:8080/hello | jq
```

## Setup networking

```
kubectl apply -f istio/gateway.yaml

// Get Gateway URL
export INGRESS_NAME=istio-ingressgateway
export INGRESS_NS=istio-system
export INGRESS_HOST=$(kubectl -n "$INGRESS_NS" get service "$INGRESS_NAME" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export INGRESS_PORT=$(kubectl -n "$INGRESS_NS" get service "$INGRESS_NAME" -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
export GATEWAY_URL=$INGRESS_HOST:$INGRESS_PORT
echo $GATEWAY_URL
```

## Access the app

```
curl -sS -H 'X-Hello-1:v2' -H 'X-Hello-2:v2' ${GATEWAY_URL}/hello | jq
// Test
for h1 in v1 v2; do for h2 in v1 v2; do curl -sS -H "X-Hello-1:${h1}" -H "X-Hello-2:${h2}" ${GATEWAY_URL}/hello; done; done
```

## Telemetry

Logs from app container

```
kubectl logs -f "$(kubectl get pod -l app=hello-2,version=v1 -o jsonpath='{.items[0].metadata.name}')" -c hello-2
```

Access logs from envoy sidecard

```
kubectl apply -f istio/telemetry.yaml
kubectl logs -f "$(kubectl get pod -l app=hello-2,version=v1 -o jsonpath='{.items[0].metadata.name}')" -c istio-proxy
```

# Clean up

```
minikube delete -p minikube
```
