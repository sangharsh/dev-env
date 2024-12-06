# Prepare certs for TLS
```
mkdir certs
# Generate a private key (tls.key):
openssl genrsa -out certs/tls.key 2048

# Generate a Certificate Signing Request (CSR):
openssl req -new -key certs/tls.key -out certs/tls.csr -subj "/CN=devenv-mesh-controller.devenv.svc" -config csr.conf

# Generate a self-signed certificate (tls.crt):
openssl x509 -req -in certs/tls.csr -signkey certs/tls.key -out certs/tls.crt -days 365 -extensions v3_req -extfile csr.conf

# (Optional) Verify the certificate:
openssl x509 -in certs/tls.crt -text -noout
```
# Build
```
eval $(minikube docker-env -p devenv) && docker build -t devenv-mesh-controller:latest .
```

# Kubernetes setup
```
# Create namespace
kubectl create namespace devenv

# Create secret
kubectl create secret tls devenv-mesh-controller-tls --cert=certs/tls.crt --key=certs/tls.key -n devenv

# (Optional) Verify secret
kubectl get secrets devenv-mesh-controller-tls -n devenv

# Deploy
kubectl apply -f deployment.yaml -n devenv

# Create ValidatingWebhookConfiguration with CA bundle
CA_BUNDLE=$(cat certs/tls.crt | base64 | tr -d '\n')
sed -e "s|caBundle:.*|caBundle: ${CA_BUNDLE}|" webhook-config.yaml | kubectl apply -f -

# ClusterRole and ClusterRoleBinding
kubectl apply -f clusterrole.yaml
kubectl apply -f clusterrolebinding.yaml
```

## Redeploy
```
eval $(minikube docker-env -p devenv) && docker build -t devenv-mesh-controller:latest . && kubectl delete deployment.apps/devenv-mesh-controller -n devenv --ignore-not-found=true && kubectl apply -f deployment.yaml -n devenv

eval $(minikube docker-env -p devenv) && docker build -t devenv-mesh-controller:latest . && kubectl delete pod -l app=devenv-mesh-controller -n devenv
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
for h1 in baseline v2; do for h2 in baseline v2; do curl -sS -H "baggage: overrides=hello-1:${h1}hello-2:${h2}" ${GATEWAY_URL}/hello; done; done
```

# Clear disk usage
```
eval $(minikube docker-env -p devenv)
docker system df
docker system prune
```
