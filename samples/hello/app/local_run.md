Note: Routing to different versions of app happens with istio.
This local run is without kubernetes and istio.
Hence, header based routing does not happen in this setup

## Run locally

Use [air](https://github.com/air-verse/air) for live reload

```
go install github.com/air-verse/air@latest
```

Run multiple instances of the go app e.g.

```
UPSTREAM_HOST=localhost:8081 air
```

```
UPSTREAM_HOST=localhost:8082 PORT=8081 MESSAGE="hello 8081" air
```

```
PORT=8082 air
```

## Call service

```
curl -sS localhost:8080/hello | jq
```
