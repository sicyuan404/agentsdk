package store

import (
	"context"

	"github.com/wordflowlab/agentsdk/pkg/types"
)

// Store 持久化存储接口
type Store interface {
	// SaveMessages 保存消息列表
	SaveMessages(ctx context.Context, agentID string, messages []types.Message) error

	// LoadMessages 加载消息列表
	LoadMessages(ctx context.Context, agentID string) ([]types.Message, error)

	// SaveToolCallRecords 保存工具调用记录
	SaveToolCallRecords(ctx context.Context, agentID string, records []types.ToolCallRecord) error

	// LoadToolCallRecords 加载工具调用记录
	LoadToolCallRecords(ctx context.Context, agentID string) ([]types.ToolCallRecord, error)

	// SaveSnapshot 保存快照
	SaveSnapshot(ctx context.Context, agentID string, snapshot types.Snapshot) error

	// LoadSnapshot 加载快照
	LoadSnapshot(ctx context.Context, agentID string, snapshotID string) (*types.Snapshot, error)

	// ListSnapshots 列出快照
	ListSnapshots(ctx context.Context, agentID string) ([]types.Snapshot, error)

	// SaveInfo 保存Agent元信息
	SaveInfo(ctx context.Context, agentID string, info types.AgentInfo) error

	// LoadInfo 加载Agent元信息
	LoadInfo(ctx context.Context, agentID string) (*types.AgentInfo, error)

	// SaveTodos 保存Todo列表
	SaveTodos(ctx context.Context, agentID string, todos interface{}) error

	// LoadTodos 加载Todo列表
	LoadTodos(ctx context.Context, agentID string) (interface{}, error)

	// DeleteAgent 删除Agent所有数据
	DeleteAgent(ctx context.Context, agentID string) error

	// ListAgents 列出所有Agent
	ListAgents(ctx context.Context) ([]string, error)
}
