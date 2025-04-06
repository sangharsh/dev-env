# Code in context

## Objective

Currently tools seem to provide creating an isolated cluster of services with changes. While isolation is a good goal, cost lies on the other side.
Existing tools in the space: Okteto, Telepresence

Enable multiple developers/teams to deploy their set of services integrated with rest of baseline services. Send request through changed and baseline services.

## Solution

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
  - Using k8s admission controller. [admctl](admctl/README.md)
- [ ] Ensure network config is consistent with deployments
  - Use k8s operator pattern
- [ ] Way to manage updates from modified service to baseline up/downstream services. Looks tricky, need to think deeper
- [ ] Support HTTP service in other languages / stacks
- [ ] Support async, message brokers etc.
- [ ] Support non-k8s workloads
