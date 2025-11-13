<template>
  <div class="mermaid-wrapper my-6">
    <div v-if="!isRendered && !hasError" class="flex justify-center items-center py-8">
      <div class="animate-pulse text-gray-500 dark:text-gray-400">
        加载图表中...
      </div>
    </div>
    <div v-if="hasError" class="flex justify-center items-center py-8">
      <div class="text-red-500 dark:text-red-400">
        图表渲染失败，请检查语法
      </div>
    </div>
    <div v-show="isRendered" ref="containerRef" class="mermaid-container flex justify-center"></div>
    <div style="display: none;">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, useSlots } from 'vue'
import { useColorMode } from '#imports'

const containerRef = ref<HTMLElement | null>(null)
const isRendered = ref(false)
const hasError = ref(false)
const colorMode = useColorMode()
const slots = useSlots()

let mermaid: any = null
let diagramId = 0

async function renderDiagram() {
  if (!containerRef.value) return

  try {
    hasError.value = false
    isRendered.value = false

    // 获取slot内容
    const slotContent = slots.default?.()[0]
    if (!slotContent || typeof slotContent.children !== 'string') {
      throw new Error('No diagram content found')
    }

    const diagramText = slotContent.children.trim()

    // 动态导入 mermaid (避免 SSR 问题)
    if (!mermaid) {
      const module = await import('mermaid')
      mermaid = module.default
    }

    // 配置 mermaid
    const isDark = colorMode.value === 'dark'
    mermaid.initialize({
      startOnLoad: false,
      theme: isDark ? 'dark' : 'default',
      securityLevel: 'loose',
      fontFamily: 'ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
      themeVariables: isDark ? {
        primaryColor: '#34d399',
        primaryTextColor: '#fff',
        primaryBorderColor: '#34d399',
        lineColor: '#60a5fa',
        secondaryColor: '#3b82f6',
        tertiaryColor: '#1f2937'
      } : {
        primaryColor: '#10b981',
        primaryTextColor: '#fff',
        primaryBorderColor: '#10b981',
        lineColor: '#3b82f6',
        secondaryColor: '#60a5fa',
        tertiaryColor: '#f3f4f6'
      }
    })

    // 清空容器
    containerRef.value.innerHTML = ''

    // 渲染图表
    const id = `mermaid-${Date.now()}-${diagramId++}`
    const { svg } = await mermaid.render(id, diagramText)
    containerRef.value.innerHTML = svg

    isRendered.value = true
  } catch (e: any) {
    console.error('Mermaid rendering error:', e)
    hasError.value = true
  }
}

// 监听主题变化，重新渲染
watch(() => colorMode.value, () => {
  if (isRendered.value) {
    renderDiagram()
  }
})

onMounted(() => {
  // 延迟一点确保DOM完全加载
  setTimeout(() => {
    renderDiagram()
  }, 100)
})
</script>

<style scoped>
.mermaid-wrapper {
  @apply overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-4;
}

.mermaid {
  @apply m-0 bg-transparent;
}

/* 确保 mermaid 图表样式不受全局 pre 样式影响 */
.mermaid-wrapper pre {
  background: transparent !important;
  border: none !important;
  padding: 0 !important;
  margin: 0 !important;
}
</style>
