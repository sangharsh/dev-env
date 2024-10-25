# Prepare certs for TLS
```
# Generate a private key (tls.key):
openssl genrsa -out certs/tls.key 2048

# Generate a Certificate Signing Request (CSR):
openssl req -new -key certs/tls.key -out certs/tls.csr -subj "/CN=admission-controller.devenv.svc" -config csr.conf

# Generate a self-signed certificate (tls.crt):
openssl x509 -req -in certs/tls.csr -signkey certs/tls.key -out certs/tls.crt -days 365 -extensions v3_req -extfile csr.conf

# (Optional) Verify the certificate:
openssl x509 -in certs/tls.crt -text -noout
```
# Build
```
eval $(minikube docker-env -p admctl)
docker build -t admission-controller:latest .
```

# Kubernetes setup
```
# Create namespace
kubectl create namespace devenv

# Create secret
kubectl create secret tls admission-controller-tls --cert=certs/tls.crt --key=certs/tls.key -n devenv

# (Optional) Verify secret
kubectl get secrets admission-controller-tls -n devenv

# Deploy
kubectl apply -f deployment.yaml -n devenv

# Create ValidatingWebhookConfiguration with CA bundle
CA_BUNDLE=$(cat certs/tls.crt | base64 | tr -d '\n')
sed -e "s|caBundle:.*|caBundle: ${CA_BUNDLE}|" webhook-config.yaml | kubectl apply -f -

# ClusterRole and ClusterRoleBinding
kubectl apply -f clusterrole.yaml
kubectl apply -f clusterrolebinding.yaml
```

## Build && Remove && Deploy
```
docker build -t admission-controller:latest . && kubectl delete deployment.apps/admission-controller -n devenv --ignore-not-found=true && kubectl apply -f deployment.yaml -n devenv

docker build -t admission-controller:latest . && kubectl delete pod -l app=admission-controller -n devenv
```

## Test
```
kubectl delete deployment nginx-deployment --ignore-not-found=true && kubectl create deployment nginx-deployment --image=nginx:latest --replicas=1
```
