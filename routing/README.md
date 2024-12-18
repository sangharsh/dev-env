# Request routing

Request routing information is sent under `overrides` key under [baggage](https://www.w3.org/TR/baggage/) header. As per its W3C Candidate Recommendation:
> This specification defines a standard for representing and propagating a set of application-defined properties associated with a distributed request or workflow execution.


Value under `overrides` is base-64 encoded protocol buffer message.

Two `envoyfilter` are created below to intercept request at gateway entry and outbound from sidecar. Envoy filter decodes value under `overrides` and for each key, value in protobuf message, it injects a request header with name (x-`key`) and value (`value`)

TODO: Ideal design was to intercept request on sidecar_inbound but injecting header at that stage is not reflecting in routing decision for some reason

## Create envoyfilter

```
export LUA_UTILS_CODE=`cat istio/utils.lua`
kubectl delete envoyfilter decode-header-sidecar --ignore-not-found=true && cat istio/decode-header-sidecar.yaml | envsubst | kubectl apply -f -
kubectl delete envoyfilter decode-header-gateway -n istio-system --ignore-not-found=true && cat istio/decode-header-gateway.yaml | envsubst | kubectl apply -f -
```

## Test
Request should now be routed to baseline and v2 versions of both apps i.e. hello-1, hello-2
```
for h1 in x y; do for h2 in x y; do curl -sS -H "baggage: overrides=Cg0KB2hlbGxvLTESAnY${h1}Cg0KB2hlbGxvLTISAnY${h2}" ${GATEWAY_URL}/hello; done; done
```
Response:
All 4 combinations are expected in response
```
{"msg":"hello-1","response":{"host":"hello-2:8080","data":{"msg":"hello-2"}}}
{"msg":"hello-1","response":{"host":"hello-2:8080","data":{"msg":"hello-2 from v2"}}}
{"msg":"hello-1 from v2","response":{"host":"hello-2:8080","data":{"msg":"hello-2"}}}
{"msg":"hello-1 from v2","response":{"host":"hello-2:8080","data":{"msg":"hello-2 from v2"}}}
```

# TODO
1. Add test cases for lua
