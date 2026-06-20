package net

import "context"

// WebSocketServer WebSocket 服务器抽象接口
// 定义 WebSocket 服务器的生命周期管理和消息处理能力。
// 具体实现由 websocket 包提供。
type WebSocketServer interface {
	// Start 启动 WebSocket 服务器
	Start() error

	// Stop 优雅停止 WebSocket 服务器
	Stop(ctx context.Context) error

	// RegisterHandler 注册消息处理器，按消息类型分发
	RegisterHandler(messageType int, handler MessageHandler)

	// RegisterMiddleware 注册 WebSocket 中间件
	RegisterMiddleware(middleware WebSocketMiddleware)

	// Broadcast 全局广播消息到所有连接
	Broadcast(message []byte, messageType int) error

	// BroadcastToRoom 向指定房间广播消息
	BroadcastToRoom(roomID string, message []byte, messageType int) error

	// GetRoom 获取指定房间实例
	GetRoom(roomID string) Room

	// GetStats 获取服务器统计信息
	GetStats() Stats
}

// MessageHandler WebSocket 消息处理器接口
//
// 处理从客户端接收的消息。
type MessageHandler interface {
	// Handle 处理连接收到的消息
	Handle(conn Connection, message []byte) error
}

// WebSocketMiddleware WebSocket 中间件接口
//
// 提供 WebSocket 连接生命周期的拦截点。
type WebSocketMiddleware interface {
	// BeforeHandshake 握手前的拦截处理，返回错误可拒绝连接
	BeforeHandshake(r interface{}) error

	// OnConnect 连接建立后的回调
	OnConnect(conn Connection)

	// OnMessage 接收消息时的回调
	OnMessage(conn Connection, message []byte)

	// OnDisconnect 连接断开后的回调
	OnDisconnect(conn Connection)
}

// Connection WebSocket 连接抽象接口
//
// 封装 WebSocket 连接操作，支持消息发送、房间管理和元数据存取。
type Connection interface {
	// ID 返回连接唯一标识
	ID() string

	// Send 发送消息到连接
	Send(message []byte, messageType int) error

	// Close 关闭连接
	Close() error

	// JoinRoom 加入指定房间
	JoinRoom(roomID string) error

	// LeaveRoom 离开指定房间
	LeaveRoom(roomID string) error

	// GetRooms 获取已加入的房间 ID 列表
	GetRooms() []string

	// SetMetadata 设置连接元数据
	SetMetadata(key string, value interface{})

	// GetMetadata 获取连接元数据
	GetMetadata(key string) (interface{}, bool)
}

// Room WebSocket 房间抽象接口
//
// 管理房间内的连接和消息广播。
type Room interface {
	// Add 添加连接到房间
	Add(conn Connection) error

	// Remove 从房间移除指定连接
	Remove(connID string) error

	// Broadcast 向房间内广播消息，可排除指定连接
	Broadcast(message []byte, messageType int, excludeConnID string) error

	// Size 返回房间内连接数量
	Size() int

	// GetConnections 获取房间内所有连接
	GetConnections() []Connection
}

// Stats WebSocket 服务器统计信息
type Stats struct {
	TotalConnections  int   // 历史总连接数
	ActiveConnections int   // 当前活跃连接数
	RoomsCount        int   // 房间数量
	MessagesSent      int64 // 已发送消息数
	MessagesReceived  int64 // 已接收消息数
	BytesSent         int64 // 已发送字节数
	BytesReceived     int64 // 已接收字节数
}
