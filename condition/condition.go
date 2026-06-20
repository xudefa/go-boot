// Package condition 提供条件判断机制，参考 Spring Boot 的 @Conditional 注解体系。
//
// 用于在自动配置（AutoConfiguration）中控制 Bean 或配置是否生效。
// 支持基于属性、Bean 存在性、Profile、类路径等条件进行判断。
//
// 使用示例：
//
//	// 基于属性条件
//	boot.RegisterAutoConfig(&GinConfig{},
//	    condition.OnProperty("gin.enabled", "true"),
//	)
//
//	// 基于 Bean 存在条件
//	condition.OnBean("redisClient")
//	condition.OnMissingBean("customCache")
//
//	// 组合条件
//	condition.All(
//	    condition.OnProperty("db.enabled"),
//	    condition.OnClass("github.com/lib/pq"),
//	)
package condition

// Condition 条件接口
//
// 参考 Spring Boot 的 @Conditional 注解体系。
// 用于在 AutoConfiguration 中控制 Bean 是否注册。
type Condition interface {
	// Matches 评估条件是否匹配
	Matches(ctx ConditionContext) bool
	// String 返回条件的可读描述，用于日志输出
	String() string
}

// ConditionContext 条件上下文
//
// 提供条件判断所需的环境、容器、类加载和一些便捷方法。
type ConditionContext interface {
	// Environment 返回环境配置，用于读取属性值
	Environment() interface{ GetProperty(key string) (any, bool) }
	// Container 返回 DI 容器，用于检查 Bean 是否存在
	Container() interface{ Has(id string) bool }
	// ClassLoader 返回类加载器，用于检查编译依赖是否存在
	ClassLoader() interface{ HasClass(name string) bool }

	// GetBean 从容器中获取 Bean 实例
	GetBean(beanID string) (any, bool)
	// HasProperty 检查属性是否存在
	HasProperty(key string) bool
	// GetProperty 获取属性值
	GetProperty(key string) (any, bool)
}
