package cache

type RedisConfig struct {
	Enable               bool   `mapstructure:"enable" json:"enable"`                                           // 是否启动
	Address              string `mapstructure:"address" default:"localhost:3306" json:"address,omitempty"`      // redis地址 (默认localhost:6379)
	Username             string `mapstructure:"username" json:"username,omitempty"`                             // redis 用户名
	Password             string `mapstructure:"password" json:"password,omitempty"`                             // redis密码
	DB                   int    `mapstructure:"db" json:"db,omitempty"`                                         // redis 库序号，默认0
	PoolSize             int    `mapstructure:"pool-size" json:"pool_size,omitempty"`                           // 连接池大小
	MaxActiveConnections int    `mapstructure:"max-active-connections" json:"max_active_connections,omitempty"` // 最大活跃连接数
	MinIdleConnections   int    `mapstructure:"min-idle-connections" json:"min_idle_connections,omitempty"`     // 最小空闲连接数
	UseCluster           bool   `mapstructure:"use-cluster" json:"use_cluster"`                                 // 是否使用集群模式
}
