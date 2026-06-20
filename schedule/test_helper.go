// Package schedule 定时任务调度框架的测试辅助文件
package schedule

import (
	"errors"

	"github.com/xudefa/go-boot/core"
)

type GetOption = core.GetOption
type BuilderOption = core.BuilderOption
type InvokeOption = core.InvokeOption

// MockContainer 模拟容器用于测试
type MockContainer struct {
	beams map[string]interface{}
}

func (m *MockContainer) Get(id string, opts ...GetOption) (interface{}, error) {
	if beam, exists := m.beams[id]; exists {
		return beam, nil
	}
	return nil, errors.New("bean not found")
}

func (m *MockContainer) ListBeans() []core.BeanInfo {
	var result []core.BeanInfo
	for id := range m.beams {
		result = append(result, core.BeanInfo{
			ID:   id,
			Type: "mock",
		})
	}
	return result
}

func (m *MockContainer) Register(beanID string, builders ...BuilderOption) error {
	if m.beams == nil {
		m.beams = make(map[string]interface{})
	}

	def := &core.BeanDefinition{
		Scope: core.SingletonScope,
	}

	for _, builder := range builders {
		if err := builder(def); err != nil {
			return err
		}
	}

	if def.Instance != nil {
		m.beams[beanID] = def.Instance
	} else if def.Factory != nil {
		instance, err := def.Factory(m)
		if err != nil {
			return err
		}
		m.beams[beanID] = instance
	}

	return nil
}

func (m *MockContainer) Inject(target interface{}) error {
	return errors.New("not implemented")
}

func (m *MockContainer) GetAll(beanType interface{}) ([]interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockContainer) Invoke(fn interface{}, opts ...InvokeOption) ([]interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockContainer) Has(beanID string) bool {
	_, exists := m.beams[beanID]
	return exists
}

func (m *MockContainer) Remove(beanID string) error {
	return errors.New("not implemented")
}

func (m *MockContainer) Close() error {
	return errors.New("not implemented")
}

func (m *MockContainer) CreateBean(beanID string) (interface{}, error) {
	return m.Get(beanID)
}
