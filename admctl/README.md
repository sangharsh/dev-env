# Prepare certs for TLS
```
mkdir admctl/certs
# Generate a private key (tls.key):
openssl genrsa -out admctl/certs/tls.key 2048

# Generate a Certificate Signing Request (CSR):
openssl req -new -key admctl/certs/tls.key -out admctl/certs/tls.csr -subj "/CN=devenv-mesh-controller.devenv.svc" -config admctl/csr.conf

# Generate a self-signed certificate (tls.crt):
openssl x509 -req -in admctl/certs/tls.csr -signkey admctl/certs/tls.key -out admctl/certs/tls.crt -days 365 -extensions v3_req -extfile admctl/csr.conf

# (Optional) Verify the certificate:
openssl x509 -in admctl/certs/tls.crt -text -noout
```
# Build
```
eval $(minikube docker-env -p devenv) && docker build -t devenv-mesh-controller:latest ./admctl
```

# Kubernetes setup
```
# Create namespace
kubectl create namespace devenv

# Create secret
kubectl create secret tls devenv-mesh-controller-tls --cert=admctl/certs/tls.crt --key=admctl/certs/tls.key -n devenv

# (Optional) Verify secret
kubectl get secrets devenv-mesh-controller-tls -n devenv

# Deploy
kubectl apply -f admctl/deployment.yaml -n devenv

# Create ValidatingWebhookConfiguration with CA bundle
CA_BUNDLE=$(cat admctl/certs/tls.crt | base64 | tr -d '\n')
sed -e "s|caBundle:.*|caBundle: ${CA_BUNDLE}|" admctl/webhook-config.yaml | kubectl apply -f -

# ClusterRole and ClusterRoleBinding
kubectl apply -f admctl/clusterrole.yaml
kubectl apply -f admctl/clusterrolebinding.yaml
```

## Redeploy
```
eval $(minikube docker-env -p devenv) && docker build -t devenv-mesh-controller:latest ./admctl && kubectl delete deployment.apps/devenv-mesh-controller -n devenv --ignore-not-found=true && kubectl apply -f admctl/deployment.yaml -n devenv

eval $(minikube docker-env -p devenv) && docker build -t devenv-mesh-controller:latest ./admctl && kubectl delete pod -l app=devenv-mesh-controller -n devenv
```

## Test
```
kubectl delete deployment nginx-deployment --ignore-not-found=true && kubectl create deployment nginx-deployment --image=nginx:latest --replicas=1
```

# Deploy v2 apps
```
kubectl delete deployment hello-1-v2 --ignore-not-found=true && kubectl apply -f istio/hello-1-v2.yaml
kubectl delete deployment hello-2-v2 --ignore-not-found=true && kubectl apply -f istio/hello-2-v2.yaml
```

## Test
```
curl -sS ${GATEWAY_URL}/hello
```
> {"msg":"hello-1","response":{"data":{"msg":"hello-2"}}}

```
curl -sS -H "x-hello-1: v2" ${GATEWAY_URL}/hello
```
> {"msg":"hello-1 from v2","response":{"data":{"msg":"hello-2"}}}

# Clear disk usage
```
eval $(minikube docker-env -p devenv)
docker system df
docker system prune
```
