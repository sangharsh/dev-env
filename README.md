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

## Checkin

Done

1. Hello service
1. Routing via istio, header based
1. Context propagation

ToDo

1. Context propagation should be supported by third party services as well. HTTP Baggage is intended for this.
   1. Single header like "overrides"
1. Network layer should be able to parse baggage to act on it. User should be able to create baggage to send request
1. Add languages. Python, node, Java
1. VM, infra pieces, async etc.
