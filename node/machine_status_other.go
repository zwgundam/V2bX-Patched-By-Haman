//go:build !linux

package node

import (
	"fmt"
	"runtime"

	"github.com/MoeclubM/V2bX/api/panel"
)

func newMachineStatusFunc() (func() (*panel.MachineStatus, error), error) {
	return nil, fmt.Errorf("machine status report is not supported on %s", runtime.GOOS)
}
