# TODO
* [] Build watcher updates when the secret changes.
* [] Move to cobra arguments so we can sepearate the coral components.
  * [] I think we need at least a controller, builder, and sync command.
* [] Create a new devtools repo to include:
  * [] Generalized cert generation scripts for webhook development
  * [] Generalized kind cluster management so we can use the same clusters cross-project
    * [] Note: the mounted directory will be an issue.  How do we do either 1) dynamic mounts or 2) do we open it up to all configured directories, or 3) open it up to an entire home folder (not a good one)?

## Tomorrow

* [x] Inject the token environment variable into the deployment.
* [] Wire up the builder watch polling and test it out.
* [] Introduction of the build image (and creation of a generalized build image for go, rust, python, and just an ordinary buildah one)
  * [] There's still some internal dialog into whether or not I want to do environment installs inside the container (slow and probably buggy) or only support the images.
  * [] This also links into how we expose build commands (as builder.yaml files in the repo).  We could potentially add "custom" build commands and "packages".
  * [] Thinking about this more, we should probably not rely on builder.yaml files inside the repo, this would tightly couple the images to these.  Instead we could add some simple generic fields/etc to expose.  Then if the user wants to make things a bit more intricate by including a build file there, they can. I'd love to just be able to define a standard go image and build.
