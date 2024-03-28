# Notes

# Worker
* Just a note that the sock location must be homogeneous across all nodes.
* Lock down the pod through RBAC policies, use a scratch image with no extra items.
Note: You can copy over /etc/ssl into the scratch image and it should work.  So use a build stage to add certs to a debian/ubuntu image, then copy.

# TODAY
* Add in secrets.  AuthConfig is in the PullImageRequest struct.
* Additional work on the readme.
* If there's time left, put together the mirroring.  Right now I have it in the Image, but I think I don't actually want that.  Create a new type called RegistryMirror.  I think that we actually create a more wholistic process here where we have a producer that pushes into a queue then a set of consumers that mirror images

# BRAINSTORM
* Only have a single list for containers?  Is there any case where I would want to have the selectors different for each filter set?

# BUG
* After a image deletion and re-apply, the monitor didn't start back up again.
* Finalizer doesn't appear to be requeuing itself, or if it is it's not running the finalizer again to remove.

# LATER
* Provide a way for coral to override annotations and force pullpolicies and selectors.  By default, have them disabled so the pre-fetch is more of a convienience feature and if the container doesn't exist on the system it pulls it so no selectors are needed.  However, there may be the case where admins will want to lock image use to those that are already available (or maybe open everything up to only-mirrored) and want to override individual settings.
* See if we can speed up image loads through local registries or shared image mounts. AWS uses a snapshotted volume that it mounts into the node so all the images are available on startup.  But I think we can still speed it up if we are hosting a local registry (may need to have an HPA attached to it to guard against scale)
* Local registry process that has a worker and a puller that pulls from external to internal registries. If the image references the local registry, the mutator will ensure that all of the container images are updated to point to the internal.
* Access to registries should include some ability to authenticate.
* Identify and add group/org to the path of containers.
* Monitor also adds a 'images' section to status with detailed state info for each tag
* Fetch workers should be able to set a state of error so we don't retry ones that have failed.  Use exponential backoff on that with a max time of 10 minutes - if an image has failed with 'failed to pull image' at least <configurable> times, set the label to error.

# CLEAN
* Consolidate all of the node selection helpers
* Consolidate hashing utility
