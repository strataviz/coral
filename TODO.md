# Notes

## CURRENT
* Create a new type called RegistryMirror.  I think that we actually create a more wholistic process here where we have a producer that pushes into a queue then a set of consumers that mirror images.  There will be a new injection rule that will transform images to use the local repository only.  That way they can keep the pull, but still restrict external pulling (and it's quicker).  Initially it requires a local registry to be available.  This may be pushed off depending on time.
* Additional work on the readme and docs.
* Dockerfiles.
* Package manifests.

## MVP
* Fix all known bugs.
* Clean up the processImage method (if time, add some testing around).
* Remove envtest from the controller in favor of the client mock.
* Move TODO items into github issues.
* Set up github actions.
* Finish and polish the README and other docs.

## BUG
* After a image deletion and re-apply, the monitor didn't start back up again (I think this is due to not shutting it down in the finalizer).
* Monitor not shutting down.
* Finalizer doesn't appear to be requeuing itself, or if it is it's not running the finalizer again to remove.

## LATER
* Standardize tests.  The layout has varied as I've gotten used to the new framework.
* Provide a way for coral to override annotations and force pullpolicies and selectors.  By default, have them disabled so the pre-fetch is more of a convienience feature and if the container doesn't exist on the system it pulls it so no selectors are needed.  However, there may be the case where admins will want to lock image use to those that are already available (or maybe open everything up to only-mirrored) and want to override individual settings.
* See if we can speed up image loads through local registries or shared image mounts. AWS uses a snapshotted volume that it mounts into the node so all the images are available on startup.  But I think we can still speed it up if we are hosting a local registry (may need to have an HPA attached to it to guard against scale)
* Local registry process that has a worker and a puller that pulls from external to internal registries. If the image references the local registry, the mutator will ensure that all of the container images are updated to point to the internal.
* Access to registries should include some ability to authenticate.
* Identify and add group/org to the path of containers.
* Monitor also adds a 'images' section to status with detailed state info for each tag
* Fetch workers should be able to set a state of error so we don't retry ones that have failed.  Use exponential backoff on that with a max time of 10 minutes - if an image has failed with 'failed to pull image' at least <configurable> times, set the label to error.
* Set up docs page in netlify.
* Container build service and uploads to internal (and potentially external) registries.  The object is to keep things local thereby negating the need to pay for private external registries.  The container build service should be relatively simple in that it just creates jobs with user provided build containers.  We can provide a base container with some standard build/deploy tools.
* Block while pulling and removing or be smart about pushing additional pull events on to the queue if we are already pulling.  The runtime handles duplicate requests, but it just seems wasteful to push them over and over.

## AGENT NOTES
* Just a note that the sock location must be homogeneous across all nodes.
* Lock down the pod through RBAC policies, use a scratch image with no extra items.
Note: You can copy over /etc/ssl into the scratch image and it should work.  So use a build stage to add certs to a debian/ubuntu image, then copy.