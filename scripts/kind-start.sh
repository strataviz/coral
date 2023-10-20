#!/usr/bin/env bash
set -o errexit

NODE_IMAGE="kindest/node:v1.28.0"

CLUSTER="$(kind get clusters 2>&1 | grep strataviz || : )"
# Only start the cluster if it doesn't exist.  I'm starting to reuse clusters so it
# may be worthwhle to pull out my sandboxing infra to something that can be reused
# across projects.
if [ "x$CLUSTER" == "x" ] ; then
cat <<EOF | kind create cluster --name=strataviz --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: $NODE_IMAGE
  extraMounts:
    - hostPath: ${PWD}
      containerPath: /app
      readOnly: true
- role: worker
  image: $NODE_IMAGE
  extraMounts:
    - hostPath: ${PWD}
      containerPath: /app
      readOnly: true
- role: worker
  image: $NODE_IMAGE
  extraMounts:
    - hostPath: ${PWD}
      containerPath: /app
      readOnly: true
- role: worker
  image: $NODE_IMAGE
  extraMounts:
    - hostPath: ${PWD}
      containerPath: /app
      readOnly: true
EOF
else
	echo "Cluster exists, skipping creation"
fi
