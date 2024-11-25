# Deploy app with istio routing

Deploy app and setup istio to route request to app version based on header

### Prerequisites

1. Kubernetes cluster e.g. [Minikube](https://minikube.sigs.k8s.io/docs/start/) (below example uses it)
1. Tool to build container image e.g. [Docker](https://www.docker.com/)
1. Service mesh like [Istio](https://istio.io/latest/docs/setup/getting-started/#download)

## Setup instructions

### Kubernetes cluster

```
minikube start -p minikube
// Run in a separate window
// Assigns an external IP for LoadBalancer services
minikube tunnel -p minikube
```

### Service mesh

```
istioctl install --skip-confirmation
kubectl label namespace default istio-injection=enabled
```

### Container image

```
eval $(minikube docker-env -p minikube)
docker build -t hello:latest -f hello/Dockerfile hello/
```

### Deploy

```
kubectl apply -f istio/deployments.yaml
// Access from within a pod
kubectl exec "$(kubectl get pod -l app=hello-2 -o jsonpath='{.items[0].metadata.name}')" -c hello-2 -- wget -q -O- hello-1:8080/hello | jq
```

### Networking

```
kubectl apply -f istio/gateway.yaml
```

Get Gateway URL

```
export INGRESS_NAME=istio-ingressgateway
export INGRESS_NS=istio-system
export INGRESS_HOST=$(kubectl -n "$INGRESS_NS" get service "$INGRESS_NAME" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export INGRESS_PORT=$(kubectl -n "$INGRESS_NS" get service "$INGRESS_NAME" -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
export GATEWAY_URL=$INGRESS_HOST:$INGRESS_PORT
echo $GATEWAY_URL
```

### Access the app

Send a request to apps via ingress gateway
```
curl -sS -H 'X-Hello-1:v2' -H 'X-Hello-2:v2' ${GATEWAY_URL}/hello | jq
```
Test all combinations
```
for h1 in baseline v2; do for h2 in baseline v2; do \
curl -sS -H "baggage: overrides=hello-1:${h1}hello-2:${h2}" ${GATEWAY_URL}/hello; \
done; done
```

### Telemetry

Logs from app container

```
kubectl logs -f "$(kubectl get pod -l app=hello-2,version=v1 -o jsonpath='{.items[0].metadata.name}')" -c hello-2
```

Access logs from envoy sidecar

```
kubectl apply -f istio/telemetry.yaml
kubectl logs -f "$(kubectl get pod -l app=hello-2,version=v1 -o jsonpath='{.items[0].metadata.name}')" -c istio-proxy
```

## Clean up

```
minikube delete -p minikube
```

## Envoy filter
Create
```
kubectl delete envoyfilter split-proto-header --ignore-not-found=true && kubectl apply -f ../istio/split-proto-header.yaml
```

Request/Response:
```
curl  -H "x-proto-data:ChUKBXVzZXJzEgwxMC4wLjAuNDI6ODAKFQoFb3JkZXISDDEwLjAuMC40Mzo4MAoNCgdoZWxsby0xEgJ2MQ==" ${GATEWAY_URL}/hello

{"msg":"hello-1","response":{"host":"hello-2:8080","data":{"msg":"hello-2"}}}
```
Istio proxy:
```
kubectl logs --tail=1 -f "$(kubectl get pod -l app=hello-1,version=baseline -o jsonpath='{.items[0].metadata.name}')" -c istio-proxy
```
Logs:
```
2024-11-25T06:28:42.898202Z     info    envoy lua external/envoy/source/extensions/filters/http/lua/lua_filter.cc:941   script log: decoded_base64:

users
     10.0.0.42:80

order
     10.0.0.43:80

hello-1v1       thread=22
2024-11-25T06:28:42.898505Z     info    envoy lua external/envoy/source/extensions/filters/http/lua/lua_filter.cc:941   script log: overrides: hello-1: v1, order: 10.0.0.43:80, users: 10.0.0.42:80,         thread=22
2024-11-25T06:28:42.898517Z     info    envoy lua external/envoy/source/extensions/filters/http/lua/lua_filter.cc:941   script log: hello-1: v1 thread=22
2024-11-25T06:28:42.898527Z     info    envoy lua external/envoy/source/extensions/filters/http/lua/lua_filter.cc:941   script log: order: 10.0.0.43:80 thread=22
2024-11-25T06:28:42.898530Z     info    envoy lua external/envoy/source/extensions/filters/http/lua/lua_filter.cc:941   script log: users: 10.0.0.42:80 thread=22
2024-11-25T06:28:42.898532Z     debug   envoy lua external/envoy/source/extensions/filters/common/lua/lua.cc:39 coroutine finished      thread=22
```

App logs:
```
kubectl logs -f "$(kubectl get pod -l app=hello-1,version=baseline -o jsonpath='{.items[0].metadata.name}')"
```
```
2024/11/25 06:28:42 handleHello headers: map[Accept:[*/*] User-Agent:[curl/8.5.0] X-Envoy-Attempt-Count:[1] X-Envoy-Internal:[true] X-Forwarded-Client-Cert:[By=spiffe://cluster.local/ns/default/sa/default;Hash=45cb4d9e1878bc19f8f84d5cc7b6aae0452df903d45bc6a00815f230a5dcec71;Subject="";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account] X-Forwarded-For:[10.244.0.1] X-Forwarded-Proto:[http] X-Hello-1:[v1] X-Order:[10.0.0.43:80] X-Proto-Data:[ChUKBXVzZXJzEgwxMC4wLjAuNDI6ODAKFQoFb3JkZXISDDEwLjAuMC40Mzo4MAoNCgdoZWxsby0xEgJ2MQ==] X-Request-Id:[10726c1a-b106-4d12-aa2b-aa29b4bdad02] X-Users:[10.0.0.42:80]]
```
