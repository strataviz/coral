package v1

import "fmt"

func (w *Watch) FullName() string {
	return fmt.Sprintf("github/%s/%s", *w.Owner, *w.Repo)
}
