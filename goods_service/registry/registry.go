package registry

import "github.com/hashicorp/consul/api"

// 面向接口开发
// 我不关心对方是什么（类型是什么），只关心对方能做什么（方法）。
// 抽象做的好，后期可以很方便的切换不同的注册中心

// Register 自定义一个注册中心的抽象（此示例不够严谨，仍然使用了consul/api库，仅做教学使用）
type Register interface {
	// 注册
	RegisterService(serviceName string, ip string, port int, tags []string) error
	// 服务发现
	ListService(serviceName string) (map[string]*api.AgentService, error)
	// 注销
	Deregister(serviceID string) error
}
