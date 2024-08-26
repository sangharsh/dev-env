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
kubectl apply -f hello/istio/deployments.yaml
// Access from within a pod
kubectl exec "$(kubectl get pod -l app=hello-2 -o jsonpath='{.items[0].metadata.name}')" -c hello-2 -- wget -q -O- hello-1:8080/hello | jq
```

## Setup networking

```
kubectl apply -f hello/istio/gateway.yaml

// Get Gateway URL
export INGRESS_NAME=istio-ingressgateway
export INGRESS_NS=istio-system
// kubectl get svc "$INGRESS_NAME" -n "$INGRESS_NS"
export INGRESS_HOST=$(kubectl -n "$INGRESS_NS" get service "$INGRESS_NAME" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export INGRESS_PORT=$(kubectl -n "$INGRESS_NS" get service "$INGRESS_NAME" -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
export GATEWAY_URL=$INGRESS_HOST:$INGRESS_PORT
echo $GATEWAY_URL
```

## Access the app

```
curl -sSv -H 'x-hello-1:v2' ${GATEWAY_URL}/hello | jq
curl -sSv ${GATEWAY_URL}/hello | jq
```

# Clean up

```
minikube delete -p minikube
```
