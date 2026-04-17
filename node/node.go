package node

import (
	"fmt"

	"github.com/MoeclubM/V2bX/api/panel"
	"github.com/MoeclubM/V2bX/conf"
	vCore "github.com/MoeclubM/V2bX/core"
)

type Node struct {
	controllers []*Controller
	machines    []*Machine
}

func New() *Node {
	return &Node{}
}

func (n *Node) Start(nodes conf.NodesConfig, core vCore.Core) error {
	runtimeNodes := nodes.RuntimeNodeConfigs()
	n.controllers = make([]*Controller, len(runtimeNodes))
	for i := range runtimeNodes {
		p, err := panel.New(&runtimeNodes[i].ApiConfig)
		if err != nil {
			n.Close()
			return err
		}
		// Register controller service
		n.controllers[i] = NewController(core, p, &runtimeNodes[i].Options)
		err = n.controllers[i].Start()
		if err != nil {
			n.Close()
			return fmt.Errorf("start node controller [%s-%s-%d] error: %s",
				runtimeNodes[i].ApiConfig.APIHost,
				runtimeNodes[i].ApiConfig.NodeType,
				runtimeNodes[i].ApiConfig.NodeID,
				err)
		}
	}
	n.machines = make([]*Machine, len(nodes.V2.Machines))
	for i := range nodes.V2.Machines {
		machine, err := NewMachine(core, &nodes.V2.Machines[i])
		if err != nil {
			n.Close()
			return err
		}
		err = machine.Start()
		if err != nil {
			n.Close()
			return fmt.Errorf("start machine controller [%s-%d] error: %s",
				nodes.V2.Machines[i].ApiConfig.APIHost,
				nodes.V2.Machines[i].ApiConfig.MachineID,
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
	for _, c := range n.controllers {
		if c == nil {
			continue
		}
		err := c.Close()
		if err != nil {
			panic(err)
		}
	}
	n.controllers = nil
}
