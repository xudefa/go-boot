package service

import (
	"github.com/xudefa/go-boot/core/container"
)

// Service 服务基类
type Service struct {
	Container container.Container
}

// GetContainer 获取容器
func (s *Service) GetContainer() container.Container {
	return s.Container
}

// SetContainer 设置容器
func (s *Service) SetContainer(c container.Container) {
	s.Container = c
}

// ServiceManager 服务管理器
type ServiceManager struct {
	container container.Container
	services  map[string]Service
}

// NewServiceManager 创建服务管理器
func NewServiceManager(c container.Container) *ServiceManager {
	return &ServiceManager{
		container: c,
		services:  make(map[string]Service),
	}
}

// Register 注册服务
func (m *ServiceManager) Register(name string, svc Service) {
	m.services[name] = svc
}

// Get 获取服务
func (m *ServiceManager) Get(name string) (Service, bool) {
	svc, exists := m.services[name]
	return svc, exists
}

// List 列出所有服务
func (m *ServiceManager) List() []string {
	names := make([]string, 0, len(m.services))
	for name := range m.services {
		names = append(names, name)
	}
	return names
}
