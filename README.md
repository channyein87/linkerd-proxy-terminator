# Linkerd Proxy Terminator

A sidecar container which terminate the linkerd proxy sidecar.\
It monitors the rest of the containers within the pod created by a job or a pod which lifecycle to be completed.\
So that pod can be terminated.

## Prerequisites

The pod requires RBAC permission to describe the pod to list the running containers.

```yaml
# Example
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: pod-viewer
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "describe"]
```

## Usage

Simply run the proxy terminator container as a sidecar along with the job pod.

```yaml
# Example
      containers:
      - name: linkerd-proxy-terminator
        image: channyein87/linkerd-proxy-terminator:latest
```

## Examples

- [simple-job](examples/simple-job)

![simple-job](docs/simple-job.gif)
