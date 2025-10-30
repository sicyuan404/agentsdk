package types

import "time"

// MessageRole 消息角色类型
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// ContentBlockType 内容块类型
type ContentBlockType string

const (
	ContentBlockTypeText       ContentBlockType = "text"
	ContentBlockTypeToolUse    ContentBlockType = "tool_use"
	ContentBlockTypeToolResult ContentBlockType = "tool_result"
)

// ContentBlock 消息内容块(接口)
type ContentBlock interface {
	Type() ContentBlockType
}

// TextBlock 文本内容块
type TextBlock struct {
	Text string `json:"text"`
}

func (t *TextBlock) Type() ContentBlockType {
	return ContentBlockTypeText
}

// ToolUseBlock 工具调用块
type ToolUseBlock struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

func (t *ToolUseBlock) Type() ContentBlockType {
	return ContentBlockTypeToolUse
}

// ToolResultBlock 工具结果块
type ToolResultBlock struct {
	ToolUseID string      `json:"tool_use_id"`
	Content   interface{} `json:"content"`
	IsError   bool        `json:"is_error,omitempty"`
}

func (t *ToolResultBlock) Type() ContentBlockType {
	return ContentBlockTypeToolResult
}

// Message AI交互消息
type Message struct {
	Role    MessageRole    `json:"role"`
	Content []ContentBlock `json:"content"`
}

// Bookmark 事件位置标记(用于续播)
type Bookmark struct {
	Seq       int64     `json:"seq"`
	Timestamp time.Time `json:"timestamp"`
}

// AgentRuntimeState Agent运行时状态
type AgentRuntimeState string

const (
	AgentStateReady   AgentRuntimeState = "READY"
	AgentStateWorking AgentRuntimeState = "WORKING"
	AgentStatePaused  AgentRuntimeState = "PAUSED"
)

// BreakpointState 断点状态(7段断点机制)
type BreakpointState string

const (
	BreakpointReady            BreakpointState = "READY"
	BreakpointPreModel         BreakpointState = "PRE_MODEL"
	BreakpointStreamingModel   BreakpointState = "STREAMING_MODEL"
	BreakpointToolPending      BreakpointState = "TOOL_PENDING"
	BreakpointAwaitingApproval BreakpointState = "AWAITING_APPROVAL"
	BreakpointPreTool          BreakpointState = "PRE_TOOL"
	BreakpointToolExecuting    BreakpointState = "TOOL_EXECUTING"
	BreakpointPostTool         BreakpointState = "POST_TOOL"
)

// ToolCallState 工具调用状态
type ToolCallState string

const (
	ToolCallStatePending          ToolCallState = "PENDING"
	ToolCallStateApprovalRequired ToolCallState = "APPROVAL_REQUIRED"
	ToolCallStateApproved         ToolCallState = "APPROVED"
	ToolCallStateExecuting        ToolCallState = "EXECUTING"
	ToolCallStateCompleted        ToolCallState = "COMPLETED"
	ToolCallStateFailed           ToolCallState = "FAILED"
	ToolCallStateDenied           ToolCallState = "DENIED"
	ToolCallStateSealed           ToolCallState = "SEALED"
)

// ToolCallApproval 工具调用审批信息
type ToolCallApproval struct {
	Required  bool                   `json:"required"`
	Decision  *string                `json:"decision,omitempty"`  // "allow" or "deny"
	DecidedBy *string                `json:"decided_by,omitempty"`
	DecidedAt *time.Time             `json:"decided_at,omitempty"`
	Note      *string                `json:"note,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// ToolCallAuditEntry 工具调用审计条目
type ToolCallAuditEntry struct {
	State     ToolCallState `json:"state"`
	Timestamp time.Time     `json:"timestamp"`
	Note      string        `json:"note,omitempty"`
}

// ToolCallRecord 工具调用完整记录
type ToolCallRecord struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Input       map[string]interface{} `json:"input"`
	State       ToolCallState          `json:"state"`
	Approval    ToolCallApproval       `json:"approval"`
	Result      interface{}            `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	IsError     bool                   `json:"is_error,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	DurationMs  *int64                 `json:"duration_ms,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	AuditTrail  []ToolCallAuditEntry   `json:"audit_trail"`
}

// ToolCallSnapshot 工具调用快照(轻量版)
type ToolCallSnapshot struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	State        ToolCallState          `json:"state"`
	Approval     ToolCallApproval       `json:"approval"`
	Result       interface{}            `json:"result,omitempty"`
	Error        string                 `json:"error,omitempty"`
	IsError      bool                   `json:"is_error,omitempty"`
	DurationMs   *int64                 `json:"duration_ms,omitempty"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	InputPreview interface{}            `json:"input_preview,omitempty"`
	AuditTrail   []ToolCallAuditEntry   `json:"audit_trail,omitempty"`
}

// Snapshot Agent状态快照
type Snapshot struct {
	ID           string                 `json:"id"`
	Messages     []Message              `json:"messages"`
	LastSfpIndex int                    `json:"last_sfp_index"`
	LastBookmark Bookmark               `json:"last_bookmark"`
	CreatedAt    time.Time              `json:"created_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AgentStatus Agent状态信息
type AgentStatus struct {
	AgentID      string            `json:"agent_id"`
	State        AgentRuntimeState `json:"state"`
	StepCount    int               `json:"step_count"`
	LastSfpIndex int               `json:"last_sfp_index"`
	LastBookmark *Bookmark         `json:"last_bookmark,omitempty"`
	Cursor       int64             `json:"cursor"`
	Breakpoint   BreakpointState   `json:"breakpoint"`
}

// AgentInfo Agent元信息
type AgentInfo struct {
	AgentID       string                 `json:"agent_id"`
	TemplateID    string                 `json:"template_id"`
	CreatedAt     time.Time              `json:"created_at"`
	Lineage       []string               `json:"lineage"`
	ConfigVersion string                 `json:"config_version"`
	MessageCount  int                    `json:"message_count"`
	LastSfpIndex  int                    `json:"last_sfp_index"`
	LastBookmark  *Bookmark              `json:"last_bookmark,omitempty"`
	Breakpoint    *BreakpointState       `json:"breakpoint,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}
