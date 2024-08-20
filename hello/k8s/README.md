## Build

```
eval $(minikube docker-env)
docker build -t hello:latest .
```

## Deploy

```
APP_NAME=hello-0 UPSTREAM_URL=http://hello-1/hello envsubst < k8s/app-env.yaml | kubectl apply -f -
APP_NAME=hello-1 UPSTREAM_URL=http://hello-2/hello envsubst < k8s/app-env.yaml | kubectl apply -f -
APP_NAME=hello-2 UPSTREAM_URL= envsubst < k8s/app-env.yaml | kubectl apply -f -
```

## Test

```
export SERVICE_NAME=hello-0; minikube service ${SERVICE_NAME} --url
curl -s http://127.0.0.1:51575/hello | jq
```

## Clean up

```
kubectl delete service --selector="stack=hello" --namespace=default
kubectl delete deployment --selector="stack=hello" --namespace=default
```
