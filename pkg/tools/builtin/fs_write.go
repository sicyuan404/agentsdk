package builtin

import (
	"context"
	"fmt"

	"github.com/wordflowlab/agentsdk/pkg/tools"
)

// FsWriteTool 文件写入工具
type FsWriteTool struct{}

// NewFsWriteTool 创建文件写入工具
func NewFsWriteTool(config map[string]interface{}) (tools.Tool, error) {
	return &FsWriteTool{}, nil
}

func (t *FsWriteTool) Name() string {
	return "fs_write"
}

func (t *FsWriteTool) Description() string {
	return "Write content to a file in the sandbox filesystem"
}

func (t *FsWriteTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *FsWriteTool) Execute(ctx context.Context, input map[string]interface{}, tc *tools.ToolContext) (interface{}, error) {
	// 获取参数
	path, ok := input["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	content, ok := input["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content must be a string")
	}

	// 写入文件
	if err := tc.Sandbox.FS().Write(ctx, path, content); err != nil {
		return map[string]interface{}{
			"ok":    false,
			"error": fmt.Sprintf("failed to write file: %v", err),
			"recommendations": []string{
				"检查文件路径是否正确",
				"确认是否有写入权限",
				"验证磁盘空间是否充足",
			},
		}, nil
	}

	return map[string]interface{}{
		"ok":     true,
		"path":   path,
		"bytes":  len(content),
		"length": len(content),
	}, nil
}

func (t *FsWriteTool) Prompt() string {
	return `Use this tool to create or overwrite files inside the sandbox.

Guidelines:
- Paths must stay inside the sandbox root. The SDK will deny attempts to escape the workspace.
- Provide the full target contents. The previous file body will be replaced.
- Pair with fs_read when editing existing files so the FilePool can validate freshness.
- The tool returns the number of bytes written for auditing purposes.

Safety/Limitations:
- File freshness validation ensures you don't overwrite externally modified files.
- Large file writes are allowed but may impact performance.`
}
