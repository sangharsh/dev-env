## Run

Run multiple instances of the go app e.g.

```
UPSTREAM_URL=http://localhost:8081/hello go run .
```

```
UPSTREAM_URL=http://localhost:8082/hello PORT=8081 MESSAGE="hello 8081" go run .
```

```
PORT=8082 go run .
```

## Call service

```
curl --silent localhost:8080/hello | jq
```
