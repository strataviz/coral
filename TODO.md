# Notes

# Worker
* Just a note that the sock location must be homogeneous across all nodes.
* Lock down the pod through RBAC policies, use a scratch image with no extra items.
Note: You can copy over /etc/ssl into the scratch image and it should work.  So use a build stage to add certs to a debian/ubuntu image, then copy.
* The node/kubelet will automatically clean up images: https://kubernetes.io/docs/concepts/architecture/garbage-collection/#container-image-garbage-collection.  Recommended to disable this, however this has the side effect of running the node out of memory so we'll want to watch this carefully in the sync service and use the high/low watermarks to make our own determination about how we clean up images, or just initially refuse to add if we hit the high watermark (used/unused/age/etc).  We could make this configurable and do a best effort attempt to remove image labels on the nodes if cleanup happens.

# TODAY
* First pass at the readme and commit.
* Add the watches and mutators for Deployments, Statefulsets, Pods, Jobs, etc for selectors and pull policy.
  * These are going to need to know how to create the selectors from the label hashes so move that
  functionality out to util.

# BUG
* After a image deletion and re-apply, the monitor didn't start back up again.
* Finalizer doesn't appear to be requeuing itself, or if it is it's not running the finalizer again to remove.

# LATER
* Add image gc to the workers (probably fits in the deletion process).
* See if we can speed up image loads through local registries or shared image mounts. AWS uses a snapshotted volume that it mounts into the node so all the images are available on startup.  But I think we can still speed it up if we are hosting a local registry (may need to have an HPA attached to it to guard against scale)
* Local registry process that has a worker and a puller that pulls from external to internal registries. If the image references the local registry, the mutator will ensure that all of the container images are updated to point to the internal.
* Access to registries should include some ability to authenticate.
* Identify and add group/org to the path of containers.
* Monitor also adds a 'images' section to status with detailed state info for each tag
* Fetch workers should be able to set a state of error so we don't retry ones that have failed.  Use exponential backoff on that with a max time of 10 minutes - if an image has failed with 'failed to pull image' at least <configurable> times, set the label to error.

# CLEAN
* Consolidate all of the node selection helpers
* Consolidate hashing utility
