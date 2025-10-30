package provider

import (
	"context"

	"github.com/wordflowlab/agentsdk/pkg/types"
)

// StreamChunk 流式响应块
type StreamChunk struct {
	Type  string      // "content_block_start", "content_block_delta", "content_block_stop", "message_delta"
	Index int         // 内容块索引
	Delta interface{} // 增量数据
	Usage *TokenUsage // Token使用情况
}

// TokenUsage Token使用统计
type TokenUsage struct {
	InputTokens  int64
	OutputTokens int64
}

// StreamOptions 流式请求选项
type StreamOptions struct {
	Tools       []ToolSchema
	MaxTokens   int
	Temperature float64
	System      string
}

// ToolSchema 工具Schema
type ToolSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// Provider 模型提供商接口
type Provider interface {
	// Stream 流式对话
	Stream(ctx context.Context, messages []types.Message, opts *StreamOptions) (<-chan StreamChunk, error)

	// Config 返回配置
	Config() *types.ModelConfig

	// Close 关闭连接
	Close() error
}

// Factory 模型提供商工厂
type Factory interface {
	Create(config *types.ModelConfig) (Provider, error)
}
