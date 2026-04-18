package node

import (
	"fmt"

	"github.com/MoeclubM/V2bX/conf"
	vCore "github.com/MoeclubM/V2bX/core"
)

type Node struct {
	machines []*Machine
}

func New() *Node {
	return &Node{}
}

func (n *Node) Start(machines []conf.MachineConfig, core vCore.Core) error {
	n.machines = make([]*Machine, len(machines))
	for i := range machines {
		machine, err := NewMachine(core, &machines[i])
		if err != nil {
			n.Close()
			return err
		}
		err = machine.Start()
		if err != nil {
			n.Close()
			return fmt.Errorf("start machine controller [%s-%d] error: %s",
				machines[i].ApiConfig.APIHost,
				machines[i].ApiConfig.MachineID,
				err)
		}
		n.machines[i] = machine
	}
	return nil
}

func (n *Node) Close() {
	for _, m := range n.machines {
		if m == nil {
			continue
		}
		m.Close()
	}
	n.machines = nil
}
