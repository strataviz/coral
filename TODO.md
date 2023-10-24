# TODO
* [] Build watcher updates when the secret changes.
* [] Move to cobra arguments so we can sepearate the coral components.
  * [] I think we need at least a controller, builder, and sync command.
* [] Create a new devtools repo to include:
  * [] Generalized cert generation scripts for webhook development
  * [] Generalized kind cluster management so we can use the same clusters cross-project
    * [] Note: the mounted directory will be an issue.  How do we do either 1) dynamic mounts or 2) do we open it up to all configured directories, or 3) open it up to an entire home folder (not a good one)?
