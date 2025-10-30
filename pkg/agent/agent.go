package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wordflowlab/agentsdk/pkg/events"
	"github.com/wordflowlab/agentsdk/pkg/provider"
	"github.com/wordflowlab/agentsdk/pkg/sandbox"
	"github.com/wordflowlab/agentsdk/pkg/tools"
	"github.com/wordflowlab/agentsdk/pkg/types"
)

// Agent AI代理
type Agent struct {
	// 基础配置
	id       string
	template *types.AgentTemplateDefinition
	config   *types.AgentConfig
	deps     *Dependencies

	// 核心组件
	eventBus   *events.EventBus
	provider   provider.Provider
	sandbox    sandbox.Sandbox
	executor   *tools.Executor
	toolMap    map[string]tools.Tool

	// 状态管理
	mu             sync.RWMutex
	state          types.AgentRuntimeState
	breakpoint     types.BreakpointState
	messages       []types.Message
	toolRecords    map[string]*types.ToolCallRecord
	stepCount      int
	lastSfpIndex   int
	lastBookmark   *types.Bookmark
	createdAt      time.Time

	// 权限管理
	pendingPermissions map[string]chan string // callID -> decision channel

	// 控制信号
	stopCh chan struct{}
}

// Create 创建新Agent
func Create(ctx context.Context, config *types.AgentConfig, deps *Dependencies) (*Agent, error) {
	// 生成AgentID
	if config.AgentID == "" {
		config.AgentID = generateAgentID()
	}

	// 获取模板
	template, err := deps.TemplateRegistry.Get(config.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}

	// 创建Provider
	modelConfig := config.ModelConfig
	if modelConfig == nil && template.Model != "" {
		modelConfig = &types.ModelConfig{
			Provider: "anthropic",
			Model:    template.Model,
		}
	}
	if modelConfig == nil {
		return nil, fmt.Errorf("model config is required")
	}

	prov, err := deps.ProviderFactory.Create(modelConfig)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	// 创建Sandbox
	sandboxConfig := config.Sandbox
	if sandboxConfig == nil {
		sandboxConfig = &types.SandboxConfig{
			Kind:    types.SandboxKindLocal,
			WorkDir: ".",
		}
	}

	sb, err := deps.SandboxFactory.Create(sandboxConfig)
	if err != nil {
		return nil, fmt.Errorf("create sandbox: %w", err)
	}

	// 创建工具执行器
	executor := tools.NewExecutor(tools.ExecutorConfig{
		MaxConcurrency: 3,
		DefaultTimeout: 60 * time.Second,
	})

	// 解析工具列表
	toolNames := config.Tools
	if toolNames == nil {
		// 使用模板的工具列表
		if toolsVal, ok := template.Tools.([]string); ok {
			toolNames = toolsVal
		} else if template.Tools == "*" {
			toolNames = deps.ToolRegistry.List()
		}
	}

	// 创建工具实例
	toolMap := make(map[string]tools.Tool)
	for _, name := range toolNames {
		tool, err := deps.ToolRegistry.Create(name, nil)
		if err != nil {
			continue // 忽略未注册的工具
		}
		toolMap[name] = tool
	}

	// 创建Agent
	agent := &Agent{
		id:                 config.AgentID,
		template:           template,
		config:             config,
		deps:               deps,
		eventBus:           events.NewEventBus(),
		provider:           prov,
		sandbox:            sb,
		executor:           executor,
		toolMap:            toolMap,
		state:              types.AgentStateReady,
		breakpoint:         types.BreakpointReady,
		messages:           []types.Message{},
		toolRecords:        make(map[string]*types.ToolCallRecord),
		pendingPermissions: make(map[string]chan string),
		createdAt:          time.Now(),
		stopCh:             make(chan struct{}),
	}

	// 初始化
	if err := agent.initialize(ctx); err != nil {
		return nil, fmt.Errorf("initialize agent: %w", err)
	}

	return agent, nil
}

// initialize 初始化Agent
func (a *Agent) initialize(ctx context.Context) error {
	// 从Store加载状态
	messages, err := a.deps.Store.LoadMessages(ctx, a.id)
	if err == nil && len(messages) > 0 {
		a.messages = messages
	}

	toolRecords, err := a.deps.Store.LoadToolCallRecords(ctx, a.id)
	if err == nil {
		for _, record := range toolRecords {
			a.toolRecords[record.ID] = &record
		}
	}

	// 保存Agent信息
	info := types.AgentInfo{
		AgentID:       a.id,
		TemplateID:    a.template.ID,
		CreatedAt:     a.createdAt,
		Lineage:       []string{},
		ConfigVersion: "v1.0.0",
		MessageCount:  len(a.messages),
	}

	return a.deps.Store.SaveInfo(ctx, a.id, info)
}

// ID 返回AgentID
func (a *Agent) ID() string {
	return a.id
}

// Send 发送消息
func (a *Agent) Send(ctx context.Context, text string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 创建用户消息
	message := types.Message{
		Role: types.MessageRoleUser,
		Content: []types.ContentBlock{
			&types.TextBlock{Text: text},
		},
	}

	a.messages = append(a.messages, message)
	a.stepCount++

	// 持久化
	if err := a.deps.Store.SaveMessages(ctx, a.id, a.messages); err != nil {
		return fmt.Errorf("save messages: %w", err)
	}

	// 触发处理
	go a.processMessages(ctx)

	return nil
}

// Chat 同步对话(阻塞式)
func (a *Agent) Chat(ctx context.Context, text string) (*types.CompleteResult, error) {
	// 发送消息
	if err := a.Send(ctx, text); err != nil {
		return nil, err
	}

	// 等待完成
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			a.mu.RLock()
			state := a.state
			a.mu.RUnlock()

			if state == types.AgentStateReady {
				// 提取最后的助手回复
				a.mu.RLock()
				defer a.mu.RUnlock()

				var text string
				for i := len(a.messages) - 1; i >= 0; i-- {
					if a.messages[i].Role == types.MessageRoleAssistant {
						for _, block := range a.messages[i].Content {
							if tb, ok := block.(*types.TextBlock); ok {
								text = tb.Text
								break
							}
						}
						break
					}
				}

				return &types.CompleteResult{
					Status: "ok",
					Text:   text,
					Last:   a.lastBookmark,
				}, nil
			}
		}
	}
}

// Subscribe 订阅事件
func (a *Agent) Subscribe(channels []types.AgentChannel, opts *types.SubscribeOptions) <-chan types.AgentEventEnvelope {
	return a.eventBus.Subscribe(channels, opts)
}

// Status 获取状态
func (a *Agent) Status() *types.AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return &types.AgentStatus{
		AgentID:      a.id,
		State:        a.state,
		StepCount:    a.stepCount,
		LastSfpIndex: a.lastSfpIndex,
		LastBookmark: a.lastBookmark,
		Cursor:       a.eventBus.GetCursor(),
		Breakpoint:   a.breakpoint,
	}
}

// Close 关闭Agent
func (a *Agent) Close() error {
	close(a.stopCh)

	if err := a.sandbox.Dispose(); err != nil {
		return err
	}

	return a.provider.Close()
}

// generateAgentID 生成AgentID
func generateAgentID() string {
	return "agt:" + uuid.New().String()
}
