#!/bin/bash

# Sample commands to patch vs and dr as needed by admission controller
# Not in use
kubectl patch virtualservice hello-1 --type json -p '
[
  {
    "op": "add",
    "path": "/spec/http/0",
    "value": {
      "match": [
        {
          "headers": {
            "baggage": {
              "regex": ".*hello-1:v2.*"
            }
          }
        }
      ],
      "route": [
        {
          "destination": {
            "host": "hello-1",
            "subset": "v2"
          }
        }
      ]
    }
  }
]'

kubectl patch destinationrule hello-1 --type json -p '
[
  {
    "op": "add",
    "path": "/spec/subsets/0",
    "value": {
      "name": "v2",
      "labels": {
        "version": "v2"
      }
    }
  }
]'

kubectl patch virtualservice hello-2 --type json -p '
[
  {
    "op": "add",
    "path": "/spec/http/0",
    "value": {
      "match": [
        {
          "headers": {
            "baggage": {
              "regex": ".*hello-2:v2.*"
            }
          }
        }
      ],
      "route": [
        {
          "destination": {
            "host": "hello-2",
            "subset": "v2"
          }
        }
      ]
    }
  }
]'

kubectl patch destinationrule hello-2 --type json -p '
[
  {
    "op": "add",
    "path": "/spec/subsets/0",
    "value": {
      "name": "v2",
      "labels": {
        "version": "v2"
      }
    }
  }
]'
