## Run

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
curl --silent localhost:8080/hello | jq
```
