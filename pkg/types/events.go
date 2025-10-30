package types

import "time"

// AgentChannel 事件通道类型
type AgentChannel string

const (
	ChannelProgress AgentChannel = "progress"
	ChannelControl  AgentChannel = "control"
	ChannelMonitor  AgentChannel = "monitor"
)

// EventType 事件类型基础接口
type EventType interface {
	Channel() AgentChannel
	EventType() string
}

// AgentEventEnvelope 事件封装(带Bookmark)
type AgentEventEnvelope struct {
	Cursor   int64       `json:"cursor"`
	Bookmark Bookmark    `json:"bookmark"`
	Event    interface{} `json:"event"`
}

// ===================
// Progress Channel Events
// ===================

// ProgressThinkChunkStartEvent 思考块开始事件
type ProgressThinkChunkStartEvent struct {
	Step int `json:"step"`
}

func (e *ProgressThinkChunkStartEvent) Channel() AgentChannel { return ChannelProgress }
func (e *ProgressThinkChunkStartEvent) EventType() string     { return "think_chunk_start" }

// ProgressThinkChunkEvent 思考块内容事件
type ProgressThinkChunkEvent struct {
	Step  int    `json:"step"`
	Delta string `json:"delta"`
}

func (e *ProgressThinkChunkEvent) Channel() AgentChannel { return ChannelProgress }
func (e *ProgressThinkChunkEvent) EventType() string     { return "think_chunk" }

// ProgressThinkChunkEndEvent 思考块结束事件
type ProgressThinkChunkEndEvent struct {
	Step int `json:"step"`
}

func (e *ProgressThinkChunkEndEvent) Channel() AgentChannel { return ChannelProgress }
func (e *ProgressThinkChunkEndEvent) EventType() string     { return "think_chunk_end" }

// ProgressTextChunkStartEvent 文本块开始事件
type ProgressTextChunkStartEvent struct {
	Step int `json:"step"`
}

func (e *ProgressTextChunkStartEvent) Channel() AgentChannel { return ChannelProgress }
func (e *ProgressTextChunkStartEvent) EventType() string     { return "text_chunk_start" }

// ProgressTextChunkEvent 文本块内容事件
type ProgressTextChunkEvent struct {
	Step  int    `json:"step"`
	Delta string `json:"delta"`
}

func (e *ProgressTextChunkEvent) Channel() AgentChannel { return ChannelProgress }
func (e *ProgressTextChunkEvent) EventType() string     { return "text_chunk" }

// ProgressTextChunkEndEvent 文本块结束事件
type ProgressTextChunkEndEvent struct {
	Step int    `json:"step"`
	Text string `json:"text"`
}

func (e *ProgressTextChunkEndEvent) Channel() AgentChannel { return ChannelProgress }
func (e *ProgressTextChunkEndEvent) EventType() string     { return "text_chunk_end" }

// ProgressToolStartEvent 工具开始执行事件
type ProgressToolStartEvent struct {
	Call ToolCallSnapshot `json:"call"`
}

func (e *ProgressToolStartEvent) Channel() AgentChannel { return ChannelProgress }
func (e *ProgressToolStartEvent) EventType() string     { return "tool:start" }

// ProgressToolEndEvent 工具执行结束事件
type ProgressToolEndEvent struct {
	Call ToolCallSnapshot `json:"call"`
}

func (e *ProgressToolEndEvent) Channel() AgentChannel { return ChannelProgress }
func (e *ProgressToolEndEvent) EventType() string     { return "tool:end" }

// ProgressToolErrorEvent 工具执行错误事件
type ProgressToolErrorEvent struct {
	Call  ToolCallSnapshot `json:"call"`
	Error string           `json:"error"`
}

func (e *ProgressToolErrorEvent) Channel() AgentChannel { return ChannelProgress }
func (e *ProgressToolErrorEvent) EventType() string     { return "tool:error" }

// ProgressDoneEvent 单轮完成事件
type ProgressDoneEvent struct {
	Step   int    `json:"step"`
	Reason string `json:"reason"` // "completed" or "interrupted"
}

func (e *ProgressDoneEvent) Channel() AgentChannel { return ChannelProgress }
func (e *ProgressDoneEvent) EventType() string     { return "done" }

// ===================
// Control Channel Events
// ===================

// RespondFunc 审批响应回调函数
type RespondFunc func(decision string, note string) error

// ControlPermissionRequiredEvent 权限请求事件
type ControlPermissionRequiredEvent struct {
	Call    ToolCallSnapshot `json:"call"`
	Respond RespondFunc      `json:"-"` // 不序列化回调函数
}

func (e *ControlPermissionRequiredEvent) Channel() AgentChannel { return ChannelControl }
func (e *ControlPermissionRequiredEvent) EventType() string     { return "permission_required" }

// ControlPermissionDecidedEvent 权限决策事件
type ControlPermissionDecidedEvent struct {
	CallID    string `json:"call_id"`
	Decision  string `json:"decision"` // "allow" or "deny"
	DecidedBy string `json:"decided_by"`
	Note      string `json:"note,omitempty"`
}

func (e *ControlPermissionDecidedEvent) Channel() AgentChannel { return ChannelControl }
func (e *ControlPermissionDecidedEvent) EventType() string     { return "permission_decided" }

// ===================
// Monitor Channel Events
// ===================

// MonitorStateChangedEvent 状态变更事件
type MonitorStateChangedEvent struct {
	State AgentRuntimeState `json:"state"`
}

func (e *MonitorStateChangedEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorStateChangedEvent) EventType() string     { return "state_changed" }

// MonitorStepCompleteEvent 步骤完成事件
type MonitorStepCompleteEvent struct {
	Step       int   `json:"step"`
	DurationMs int64 `json:"duration_ms,omitempty"`
}

func (e *MonitorStepCompleteEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorStepCompleteEvent) EventType() string     { return "step_complete" }

// MonitorErrorEvent 错误事件
type MonitorErrorEvent struct {
	Severity string                 `json:"severity"` // "info", "warn", "error"
	Phase    string                 `json:"phase"`    // "model", "tool", "system", "lifecycle"
	Message  string                 `json:"message"`
	Detail   map[string]interface{} `json:"detail,omitempty"`
}

func (e *MonitorErrorEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorErrorEvent) EventType() string     { return "error" }

// MonitorTokenUsageEvent Token使用统计事件
type MonitorTokenUsageEvent struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
}

func (e *MonitorTokenUsageEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorTokenUsageEvent) EventType() string     { return "token_usage" }

// MonitorToolExecutedEvent 工具执行完成事件
type MonitorToolExecutedEvent struct {
	Call ToolCallSnapshot `json:"call"`
}

func (e *MonitorToolExecutedEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorToolExecutedEvent) EventType() string     { return "tool_executed" }

// MonitorAgentResumedEvent Agent恢复事件
type MonitorAgentResumedEvent struct {
	Strategy string             `json:"strategy"` // "crash" or "manual"
	Sealed   []ToolCallSnapshot `json:"sealed"`
}

func (e *MonitorAgentResumedEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorAgentResumedEvent) EventType() string     { return "agent_resumed" }

// MonitorBreakpointChangedEvent 断点变更事件
type MonitorBreakpointChangedEvent struct {
	Previous  BreakpointState `json:"previous"`
	Current   BreakpointState `json:"current"`
	Timestamp time.Time       `json:"timestamp"`
}

func (e *MonitorBreakpointChangedEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorBreakpointChangedEvent) EventType() string     { return "breakpoint_changed" }

// MonitorFileChangedEvent 文件变更事件
type MonitorFileChangedEvent struct {
	Path  string    `json:"path"`
	Mtime time.Time `json:"mtime"`
}

func (e *MonitorFileChangedEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorFileChangedEvent) EventType() string     { return "file_changed" }

// MonitorReminderSentEvent 系统提醒事件
type MonitorReminderSentEvent struct {
	Category string `json:"category"` // "file", "todo", "security", "performance", "general"
	Content  string `json:"content"`
}

func (e *MonitorReminderSentEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorReminderSentEvent) EventType() string     { return "reminder_sent" }

// MonitorContextCompressionEvent 上下文压缩事件
type MonitorContextCompressionEvent struct {
	Phase   string  `json:"phase"` // "start" or "end"
	Summary string  `json:"summary,omitempty"`
	Ratio   float64 `json:"ratio,omitempty"`
}

func (e *MonitorContextCompressionEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorContextCompressionEvent) EventType() string     { return "context_compression" }

// MonitorSchedulerTriggeredEvent 调度器触发事件
type MonitorSchedulerTriggeredEvent struct {
	TaskID      string    `json:"task_id"`
	Spec        string    `json:"spec"`
	Kind        string    `json:"kind"` // "steps", "time", "cron"
	TriggeredAt time.Time `json:"triggered_at"`
}

func (e *MonitorSchedulerTriggeredEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorSchedulerTriggeredEvent) EventType() string     { return "scheduler_triggered" }

// MonitorToolManualUpdatedEvent 工具手册更新事件
type MonitorToolManualUpdatedEvent struct {
	Tools     []string  `json:"tools"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *MonitorToolManualUpdatedEvent) Channel() AgentChannel { return ChannelMonitor }
func (e *MonitorToolManualUpdatedEvent) EventType() string     { return "tool_manual_updated" }
