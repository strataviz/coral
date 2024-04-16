# Notes

## NOTES
* Remember to Clean up image/repository namings in all the packages and docs.

## CURRENT
* Finish Mirror
* Move agent lists to new NormalizedList to support inexplicit names.
* Clean up spec image/repository namings in all the packages and docs.
* I like the informer method of getting the mirrors, so move monitor and agent to this instead of listing.  Should be straightforward and cut back some of the calls.  What I don't know is how much pressure that will put on the api noteably with the daemons on every node.  I can't imagine that it is any more than having all of them hit the apiserver/etc every iteration and pull down lists.
* Add the parse step from the normalization method into the validation step for repositories.  This will allow us to catch repo naming errors before admission.
* Deploy to dockerhub.
* Comments.
* Finish and polish the README and other docs.
* Package manifests and install docs.
* Move TODO items into github issues.
* Delete the TODO file ;)
* Watch secrets in the agent and mirror.  We need to be notified when updates to our referenced secrets are updated.

## MVP
* Ready after current

## BUGS
* NA

## LATER
* Support the `all` option for syncing repository tags.
* Garbage collection for local registries.
* Better monitor with dedicated workers instead of a single process per image.
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
