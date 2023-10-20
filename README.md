# coral

Coral is a build system for kubernetes environments.  It has two main components to manage the image lifecycle.

## Image Builer

Coral will watch source repositories (currently supports public and private Github repositories) and when updates occur (either based on merging or tagging) will trigger connfigurable testing and then the building of images.  Optionally, the images can be pushed to external registries that can respond to the Docker Registry V2 API.  By default, the images are stored in a local repository in the cluster.  The image builder also provides mechanisms to mirror external images to the internal registry.

## Image Synchronization

To increase startup performance coral contains a utility that will ensure that any image (all images in the local registry or configured images in external registries) are pre-fetched and are available on the nodes.  It also provides an admissions webhook that can modify the image field and image pull policy for containers to point any image to the local repository (for accessing images before synchronization has occured).
