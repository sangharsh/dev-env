Note: Routing to different versions of app happens with istio.
This local run is without kubernetes and istio.
Hence, header based routing does not happen in this setup

## Run locally

Run multiple instances of the go app e.g.

```
UPSTREAM_HOST=localhost:8081 go run .
```

```
UPSTREAM_HOST=localhost:8082 PORT=8081 MESSAGE="hello 8081" go run .
```

```
PORT=8082 go run .
```

## Call service

```
curl -sS localhost:8080/hello | jq
```
