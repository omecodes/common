package process

import (
	"github.com/omecodes/common/errors"
	"sync"
)

type Manager struct {
	sync.Mutex
	processes      map[string]*Process
	paramsWrappers []ParamsWrapper
}

func (m *Manager) AddParamsWrapper(wrapper ParamsWrapper) {
	m.Lock()
	defer m.Unlock()
	m.paramsWrappers = append(m.paramsWrappers, wrapper)
}

func (m *Manager) List() []string {
	m.Lock()
	defer m.Unlock()
	var names []string
	for name, _ := range m.processes {
		names = append(names, name)
	}
	return names
}

func (m *Manager) Register(process *Process) error {
	m.Lock()
	defer m.Unlock()
	name := process.Name()
	_, found := m.processes[process.Name()]
	if found {
		return errors.New("found existing process with the same name")
	}
	m.processes[name] = process
	return nil
}

func (m *Manager) UnRegister(name string) error {
	m.Lock()
	defer m.Unlock()
	_, found := m.processes[name]
	if !found {
		return errors.New("process not found")
	}
	delete(m.processes, name)
	return nil
}

func (m *Manager) Start(p *Params) (*Process, error) {
	m.Lock()
	defer m.Unlock()
	name := p.Name
	_, found := m.processes[name]
	if found {
		return nil, errors.New("found existing process with that name")
	}

	var err error
	for _, w := range m.paramsWrappers {
		p, err = w.Wrap(p)
		if err != nil {
			return nil, err
		}
	}
	process := NewProcess(p)
	err = process.Start()
	return process, err
}

func (m *Manager) Restart(name string) error {
	m.Lock()
	defer m.Unlock()

	p, found := m.processes[name]
	if !found {
		return errors.New("process not found")
	}
	return p.Restart()
}

func (m *Manager) Count() int {
	m.Lock()
	defer m.Unlock()
	return len(m.processes)
}

func (m *Manager) KillProcessByName(name string) error {
	m.Lock()
	defer m.Unlock()

	p, found := m.processes[name]
	if !found {
		return errors.New("not process found")
	}
	return p.Stop()
}

func (m *Manager) GetProcess(name string) (*Process, bool) {
	m.Lock()
	defer m.Unlock()
	p, found := m.processes[name]
	return p, found
}

func NewManager() *Manager {
	return &Manager{
		processes:      map[string]*Process{},
		paramsWrappers: nil,
	}
}
