package builtin

import (
	"context"
	"fmt"
	"strings"

	"github.com/wordflowlab/agentsdk/pkg/tools"
)

// FsReadTool 文件读取工具
type FsReadTool struct{}

// NewFsReadTool 创建文件读取工具
func NewFsReadTool(config map[string]interface{}) (tools.Tool, error) {
	return &FsReadTool{}, nil
}

func (t *FsReadTool) Name() string {
	return "fs_read"
}

func (t *FsReadTool) Description() string {
	return "Read file contents from the sandbox filesystem"
}

func (t *FsReadTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read",
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "Line offset to start reading from (optional)",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of lines to read (optional)",
			},
		},
		"required": []string{"path"},
	}
}

func (t *FsReadTool) Execute(ctx context.Context, input map[string]interface{}, tc *tools.ToolContext) (interface{}, error) {
	// 获取参数
	path, ok := input["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	offset := 0
	if o, ok := input["offset"].(float64); ok {
		offset = int(o)
	}

	limit := 0
	if l, ok := input["limit"].(float64); ok {
		limit = int(l)
	}

	// 读取文件
	content, err := tc.Sandbox.FS().Read(ctx, path)
	if err != nil {
		return map[string]interface{}{
			"ok":    false,
			"error": fmt.Sprintf("failed to read file: %v", err),
			"recommendations": []string{
				"检查文件路径是否正确",
				"确认文件是否存在",
				"验证是否有读取权限",
			},
		}, nil
	}

	// 分割成行
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// 应用offset和limit
	startLine := offset
	if startLine < 0 {
		startLine = 0
	}
	if startLine >= totalLines {
		return map[string]interface{}{
			"ok":        true,
			"path":      path,
			"content":   "",
			"offset":    offset,
			"limit":     limit,
			"truncated": false,
			"totalLines": totalLines,
		}, nil
	}

	endLine := totalLines
	truncated := false
	if limit > 0 {
		endLine = startLine + limit
		if endLine > totalLines {
			endLine = totalLines
		} else {
			truncated = true
		}
	}

	selectedLines := lines[startLine:endLine]
	resultContent := strings.Join(selectedLines, "\n")

	return map[string]interface{}{
		"ok":         true,
		"path":       path,
		"content":    resultContent,
		"offset":     offset,
		"limit":      limit,
		"truncated":  truncated,
		"totalLines": totalLines,
		"readLines":  len(selectedLines),
	}, nil
}

func (t *FsReadTool) Prompt() string {
	return `## fs_read - 读取文件内容

**用途**: 从沙箱文件系统读取文件内容

**参数**:
- path (必填): 文件路径
- offset (可选): 起始行号,默认0
- limit (可选): 读取行数,默认读取全部

**返回**:
- ok: 是否成功
- content: 文件内容
- truncated: 是否被截断
- totalLines: 总行数

**示例**:
` + "```json\n" + `{
  "path": "src/main.go",
  "offset": 0,
  "limit": 100
}
` + "```\n" + `

**注意事项**:
- 路径必须在沙箱工作目录内
- 大文件建议使用offset和limit分批读取
- 读取后内容会被记录到FilePool中
`
}
