package node

import (
	"fmt"
	"sync"

	"github.com/MoeclubM/V2bX/api/panel"
	"github.com/MoeclubM/V2bX/common/task"
	"github.com/MoeclubM/V2bX/conf"
	vCore "github.com/MoeclubM/V2bX/core"
	log "github.com/sirupsen/logrus"
)

type Machine struct {
	server      vCore.Core
	config      *conf.V2MachineConfig
	apiClient   *panel.Client
	controllers map[int]*Controller
	pullTask    *task.Task
	statusTask  *task.Task
	statusFunc  func() (*panel.MachineStatus, error)
	access      sync.Mutex
}

func NewMachine(server vCore.Core, config *conf.V2MachineConfig) (*Machine, error) {
	apiConfig := config.RuntimeAPIConfig(0)
	client, err := panel.New(&apiConfig)
	if err != nil {
		return nil, err
	}
	statusFunc, err := newMachineStatusFunc()
	if err != nil {
		log.WithFields(log.Fields{
			"machine_id": config.ApiConfig.MachineID,
			"err":        err,
		}).Error("Init machine status reporter failed")
	}
	return &Machine{
		server:      server,
		config:      config,
		apiClient:   client,
		controllers: make(map[int]*Controller),
		statusFunc:  statusFunc,
	}, nil
}

func (m *Machine) Start() error {
	resp, err := m.apiClient.GetMachineNodes()
	if err != nil {
		return fmt.Errorf("get machine nodes error: %s", err)
	}
	m.access.Lock()
	err = m.syncNodes(resp.Nodes, true)
	m.access.Unlock()
	if err != nil {
		return err
	}
	m.updateTasks(resp)
	return nil
}

func (m *Machine) Close() {
	m.access.Lock()
	defer m.access.Unlock()
	if m.pullTask != nil {
		m.pullTask.Close()
		m.pullTask = nil
	}
	if m.statusTask != nil {
		m.statusTask.Close()
		m.statusTask = nil
	}
	for id, controller := range m.controllers {
		if err := controller.Close(); err != nil {
			panic(err)
		}
		delete(m.controllers, id)
	}
}

func (m *Machine) syncTask() error {
	resp, err := m.apiClient.GetMachineNodes()
	if err != nil {
		log.WithFields(log.Fields{
			"machine_id": m.config.ApiConfig.MachineID,
			"err":        err,
		}).Error("Get machine nodes failed")
		return nil
	}
	m.access.Lock()
	defer m.access.Unlock()
	if err = m.syncNodes(resp.Nodes, false); err != nil {
		log.WithFields(log.Fields{
			"machine_id": m.config.ApiConfig.MachineID,
			"err":        err,
		}).Error("Sync machine nodes failed")
		return nil
	}
	m.updateTasks(resp)
	return nil
}

func (m *Machine) updateTasks(resp *panel.MachineNodesResponse) {
	pullInterval := resp.PullInterval()
	if pullInterval > 0 {
		if m.pullTask == nil {
			m.pullTask = &task.Task{
				Interval: pullInterval,
				Execute:  m.syncTask,
			}
			log.WithField("machine_id", m.config.ApiConfig.MachineID).Info("Start machine node sync")
			_ = m.pullTask.Start(false)
		} else if m.pullTask.Interval != pullInterval {
			m.pullTask.Interval = pullInterval
			m.pullTask.Close()
			_ = m.pullTask.Start(false)
		}
	} else if m.pullTask != nil {
		m.pullTask.Close()
		m.pullTask = nil
	}
	pushInterval := resp.PushInterval()
	if pushInterval > 0 && m.statusFunc != nil {
		if m.statusTask == nil {
			m.statusTask = &task.Task{
				Interval: pushInterval,
				Execute:  m.reportStatusTask,
			}
			log.WithField("machine_id", m.config.ApiConfig.MachineID).Info("Start machine status report")
			_ = m.statusTask.Start(false)
		} else if m.statusTask.Interval != pushInterval {
			m.statusTask.Interval = pushInterval
			m.statusTask.Close()
			_ = m.statusTask.Start(false)
		}
	} else if m.statusTask != nil {
		m.statusTask.Close()
		m.statusTask = nil
	}
}

func (m *Machine) reportStatusTask() error {
	if m.statusFunc == nil {
		return nil
	}
	status, err := m.statusFunc()
	if err != nil {
		log.WithFields(log.Fields{
			"machine_id": m.config.ApiConfig.MachineID,
			"err":        err,
		}).Error("Get machine status failed")
		return nil
	}
	if err = m.apiClient.ReportMachineStatus(status); err != nil {
		log.WithFields(log.Fields{
			"machine_id": m.config.ApiConfig.MachineID,
			"err":        err,
		}).Error("Report machine status failed")
	}
	return nil
}

func (m *Machine) syncNodes(nodes []panel.MachineNode, strict bool) error {
	next := make(map[int]panel.MachineNode, len(nodes))
	for i := range nodes {
		next[nodes[i].Id] = nodes[i]
	}
	started := make([]int, 0, len(nodes))
	for i := range nodes {
		if _, ok := m.controllers[nodes[i].Id]; ok {
			continue
		}
		controller, err := m.newController(nodes[i])
		if err != nil {
			if !strict {
				log.WithFields(log.Fields{
					"machine_id": m.config.ApiConfig.MachineID,
					"node_id":    nodes[i].Id,
					"err":        err,
				}).Error("Start machine node failed")
				continue
			}
			for j := range started {
				if closeErr := m.controllers[started[j]].Close(); closeErr != nil {
					panic(closeErr)
				}
				delete(m.controllers, started[j])
			}
			return err
		}
		m.controllers[nodes[i].Id] = controller
		started = append(started, nodes[i].Id)
		log.WithFields(log.Fields{
			"machine_id": m.config.ApiConfig.MachineID,
			"node_id":    nodes[i].Id,
		}).Info("Machine node started")
	}
	for id, controller := range m.controllers {
		if _, ok := next[id]; ok {
			continue
		}
		if err := controller.Close(); err != nil {
			return err
		}
		delete(m.controllers, id)
		log.WithFields(log.Fields{
			"machine_id": m.config.ApiConfig.MachineID,
			"node_id":    id,
		}).Info("Machine node removed")
	}
	return nil
}

func (m *Machine) newController(node panel.MachineNode) (*Controller, error) {
	runtimeNode := conf.NodeConfig{
		ApiConfig: m.config.RuntimeAPIConfig(node.Id),
		Options:   m.config.Options,
	}
	if runtimeNode.Options.Name != "" {
		runtimeNode.Options.Name = fmt.Sprintf("%s-%d", runtimeNode.Options.Name, node.Id)
	}
	client, err := panel.New(&runtimeNode.ApiConfig)
	if err != nil {
		return nil, err
	}
	controller := NewController(m.server, client, &runtimeNode.Options)
	if err = controller.Start(); err != nil {
		return nil, fmt.Errorf("start node controller [%s-%d-%d] error: %s",
			runtimeNode.ApiConfig.APIHost,
			runtimeNode.ApiConfig.MachineID,
			runtimeNode.ApiConfig.NodeID,
			err)
	}
	return controller, nil
}
