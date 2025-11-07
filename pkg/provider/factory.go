package provider

import (
	"fmt"

	"github.com/wordflowlab/agentsdk/pkg/types"
)

// ProviderFactory 提供商工厂接口
type ProviderFactory interface {
	Create(config *types.ModelConfig) (Provider, error)
}

// MultiProviderFactory 多提供商工厂
type MultiProviderFactory struct{}

// NewMultiProviderFactory 创建多提供商工厂
func NewMultiProviderFactory() *MultiProviderFactory {
	return &MultiProviderFactory{}
}

// Create 根据配置创建相应的提供商
func (f *MultiProviderFactory) Create(config *types.ModelConfig) (Provider, error) {
	providerType := config.Provider
	if providerType == "" {
		// 默认使用 anthropic
		providerType = "anthropic"
	}

	switch providerType {
	case "anthropic":
		return NewAnthropicProvider(config)
	case "glm", "zhipu", "bigmodel":
		return NewGLMProvider(config)
	case "deepseek":
		return NewDeepseekProvider(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}
}
