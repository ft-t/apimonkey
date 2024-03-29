package instance

import (
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/valyala/fastjson"
)

type Manager struct {
	instances map[string]*Instance
	mut       sync.Mutex
	factory   Factory
}

func NewManager(
	factory Factory,
) *Manager {
	return &Manager{
		instances: make(map[string]*Instance),
		mut:       sync.Mutex{},
		factory:   factory,
	}
}

func (m *Manager) InitInstance(ctxId string) (*Instance, error) {
	m.mut.Lock()
	defer m.mut.Unlock()
	instance, ok := m.instances[ctxId]

	if !ok {
		instance = m.factory.Create(ctxId)
		m.instances[ctxId] = instance
	}

	return instance, nil
}

func (m *Manager) StartAsync(ctxId string) error {
	m.mut.Lock()
	defer m.mut.Unlock()
	instance, ok := m.instances[ctxId]

	if !ok {
		return errors.New("instance not found")
	}

	instance.StartAsync()

	return nil
}

func (m *Manager) KeyPressed(ctxId string) error {
	m.mut.Lock()
	instance, ok := m.instances[ctxId]
	m.mut.Unlock()

	if !ok {
		return errors.New("instance not found")
	}

	if err := instance.KeyPressed(); err != nil {
		instance.ShowAlert()

		return err
	}

	return nil
}

func (m *Manager) Stop(ctxId string) error {
	m.mut.Lock()
	defer m.mut.Unlock()
	instance, ok := m.instances[ctxId]

	if !ok {
		return nil
	}

	delete(m.instances, ctxId)
	instance.Stop()

	return nil
}

func (m *Manager) SetInstanceConfig(
	ctxId string,
	payload *fastjson.Value,
) error {
	m.mut.Lock()
	defer m.mut.Unlock()
	instance, ok := m.instances[ctxId]

	if !ok {
		return errors.New("instance not found")
	}

	if err := instance.SetConfig(payload); err != nil {
		instance.ShowAlert()

		return err
	}

	return nil
}
