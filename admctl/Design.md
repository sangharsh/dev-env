
# Kubernetes control plane
1. Admission controller
1. Operator

## Admission controller
1. Trigger on deployment CREATE and DELETE
1. Is it a baseline deployment or feature
1. Check whether istio is installed
1. Find virtual service (VS) for the deployment
1. Create patch for VS
1. Check whether patch is already applied (Idempotency)
1. Apply patch
1. Is there anything to be done for DestinationRule (seems yes), Service?

## Operator
1. CRD - Schema for configuring devenv
1. CR - Values for above CRD
1. Operator - check all deployments and VS configs. Modify if needed
