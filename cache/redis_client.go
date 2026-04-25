package cache

import (
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

func NewRedisClient(config *RedisConfig) *redis.Client {
	if config.Enable && !config.UseCluster {
		clusterOptions := buildRedisOptions(config)
		return redis.NewClient(clusterOptions)
	}
	return nil
}

func NewRedisClusterClient(config *RedisConfig) *redis.ClusterClient {
	if config.Enable && config.UseCluster {
		clusterOptions := buildRedisClusterOptions(config)
		return redis.NewClusterClient(clusterOptions)
	}
	return nil
}

func buildRedisOptions(redisConfig *RedisConfig) *redis.Options {
	options := &redis.Options{
		//连接信息
		Addr:     redisConfig.Address,  // 主机名+冒号+端口
		Username: redisConfig.Username, // 用户名
		Password: redisConfig.Password, // 密码
		DB:       0,                    // redis数据库index

		//连接池容量及闲置连接数量
		PoolSize:        redisConfig.PoolSize,             // 连接池最大socket连接数，默认为10倍CPU数， 10 * runtime.NumCPU
		MaxActiveConns:  redisConfig.MaxActiveConnections, //在给定时间内，允许分配的最大连接数（当为0时，没有限制），默认为0
		MinIdleConns:    redisConfig.MinIdleConnections,   //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。
		ConnMaxLifetime: time.Minute * 60,                 //连接存活最长时长，默认为0，即不关闭存活时长较长的连接

		// LIFO 并不适合作为负载均衡算法的选择。因为 LIFO 会优先处理最近使用过的连接，这可能会导致某些服务实例负载过重，而其他的服务实例却得
		// 不到充分的利用。这种不均衡的分配会影响系统的可用性、性能和容错能力。
		PoolFIFO: true, // 采用先进先出，true为采用，false为不采用
	}
	return options
}

func buildRedisClusterOptions(redisConfig *RedisConfig) *redis.ClusterOptions {
	// 初始化redisOptions
	clusterOptions := &redis.ClusterOptions{
		//连接信息
		Addrs:    strings.Split(redisConfig.Address, ","), // 主机名+冒号+端口
		Username: redisConfig.Username,                    // 用户名
		Password: redisConfig.Password,                    // 密码
		//连接池容量及闲置连接数量
		PoolSize:        20,               // 连接池最大socket连接数，默认为10倍CPU数， 10 * runtime.NumCPU
		MaxActiveConns:  40,               //在给定时间内，允许分配的最大连接数（当为0时，没有限制），默认为0
		MinIdleConns:    2,                //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。
		ConnMaxLifetime: 60 * time.Minute, //连接存活最长时长，默认为0，即不关闭存活时长较长的连接
		// LIFO 并不适合作为负载均衡算法的选择。因为 LIFO 会优先处理最近使用过的连接，这可能会导致某些服务实例负载过重，
		// 而其他的服务实例却得不到充分的利用。这种不均衡的分配会影响系统的可用性、性能和容错能力。
		PoolFIFO: true, // 采用先进先出，true为采用，false为不采用
	}
	return clusterOptions
}
