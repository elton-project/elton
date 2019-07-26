package subsystems

import "fmt"

type RestartRequest struct {
	Reason string
}

func (rr *RestartRequest) Error() string {
	return fmt.Sprintf("RestartRequest: %s", rr.Reason)
}
