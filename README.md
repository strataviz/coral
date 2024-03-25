# Coral

Coral is the start of a set of services for kubernetes that provides a structural framework for running applications.  The first iteration provides image management tools which lets users prefetch external container images to kubernetes nodes or to internal local registries.  It allows the user to selectively target groups of nodes for images to be fetched on, and tracks image availability on the nodes.  This allows the service to provide a mutating admissions webhook which can optionally modify the node selectors in resource specifications to ensure applications are only started on nodes that have the image available.  The webhook can also optionally modify the pull policies for containers to ensure that containers do not pull external containers and only rely on the configured container images that are already present in the cluster.

Over time we expect to add additional features to Coral including:
- [ ] Fully managemed internal registries.
- [ ] A comprehensive build system that builds and stores to internal registries allowing users to bypass external registries all together.
- [ ] Client tools to allow for the triggering of build from local source directories or remote repositories, enabling a more fluid development process.
- [ ] Registry mirroring and cross-cluster image synchronization.
- [ ] Additional garbage collection and lifecycle management features for container images on both the nodes and internal registries.
 
## Installation

```
kubectl apply -k <url>
```

## Usage

### Prefetching images from external repositories directly to the nodes.

TODO

### Mirroring images from external repositories to an internal repository.

TODO

### Node selector mutations

TODO

### Enabling pull policy mutations

TODO

## Fetch workers

TODO

### Configuration

TODO

### Security concerns

The fetch workers interact with the node by mounting the runtime socket and using the Kubernetes CRI-API wrapper around the container runtime environment.  This does introduce potential attack vectors to the service and is generally discouraged.  With this in mind, we built the service to minimize the surface area exposed.

1) The fetch worker containers are built without an operating system or system utilities which do not provide any way to execute commands remotely.  We are only interacting with images through the Kubernetes CRI-API and not fetching images directly.
2) Exposed APIs are read only (currently only exposing metrics).

This should minimize the potential for abuse considerably.

## Potential issues

* Kubernetes provides internal image [https://kubernetes.io/docs/concepts/architecture/garbage-collection/#container-image-garbage-collection](garbage collection based on a series of constraints).  The fetch workers rely on both node labels that it manages and the availability of the image as reported by the node to determine whether or not to fetch it or where it is in it's lifecycle.  If the image has been expunged from the node, the fetch client will attempt to retrieve it again potentially leading to thrashing if the GC is caused by a disk pressure situation.  One possible solution would be to disable the kubelet's image garbage collection.  Coral will track disk/pid pressure situations and will not fetch new images until addressed.  This state emits metrics for monitoring and alerting, which will allow for the user to respond and remove images more intellegently. However, this only makes sense if you are managing all images through coral.  There are efforts being planned to handle this situation more gracefully.

## Development

TODO