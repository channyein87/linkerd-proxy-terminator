# Linkerd Proxy Terminator

A sidecar container which terminate the linkerd proxy sidecar.\
It monitors the rest of the containers within the pod created by a job or a pod which lifecycle to be completed.\
So that pod can be terminated.
