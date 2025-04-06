# Better testing while developing

Many times, we change and test microservices in local workstation without access to upstream/downstream services. Integration testing happens at a later point in time. Major reason for this is effort and maintanence required to manage dependent services.

[Lyft blog](https://eng.lyft.com/scaling-productivity-on-microservices-at-lyft-part-3-extending-our-envoy-mesh-with-staging-fdaafafca82f) discusses their evolution in this journey. This project draws many inspiration from part 3 of the blog. It tries to implement same functionality in a Kubernetes cluster, aiming to be useful to wider audience.

## Approach

1. All baseline services are deployed in k8s cluster.
1. Normal request make upstream calls as needed to baseline services using some sort of service mesh
1. Developers should be able to deploy their changed services using an agreed-upon label.
1. Request with a specific header use services deployed with matching label if exists otherwise baseline service.
   1. Header should be forwarded in S2S (also async, infra?(MQ, DB, SQS) etc.) calls

## Status

Done

- [x] HTTP service with a upstream HTTP call
  - hello service in golang
- [x] Context propagation
  - Using baggage header and open-telemetry SDK
- [x] Network routing based on request header
  - [Routing](routing/README.md). Used Istio, EnvoyFilter, Lua
- [x] Update network on deployments
  - Using k8s admission controller hooked to deployments. [admctl](admctl/README.md)
- [ ] Ensure network config is consistent with deployments
  - Use k8s operator pattern
- [ ] Way to manage updates from modified service to baseline up/downstream services. Looks tricky, need to think deeper
- [ ] Support HTTP service in other languages / stacks
- [ ] Support async, message brokers etc.
- [ ] Support non-k8s workloads

## Related work

Network routing
1. Telepresence
1. Mirrord
1. Signadot

In-sync deploy on cloud
1. Okteto
