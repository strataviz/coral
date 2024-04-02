package image

type ControllerErrors string

func (e ControllerErrors) Error() string {
	return string(e)
}

const (
	ErrNodesNotEmpty ControllerErrors = "nodes have not removed all images"
)
