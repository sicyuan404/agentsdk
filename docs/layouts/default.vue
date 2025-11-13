<template>
  <div class="min-h-screen bg-white dark:bg-gray-900">
    <!-- 头部导航 -->
    <header class="sticky top-0 z-50 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
      <nav class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <div class="flex items-center space-x-8">
          <NuxtLink to="/" class="text-xl font-bold text-gray-900 dark:text-white">
            AgentSDK
          </NuxtLink>
          <div class="hidden md:flex space-x-6">
            <NuxtLink to="/introduction/overview" class="text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white">
              文档
            </NuxtLink>
            <NuxtLink to="/api-reference/agent-api" class="text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white">
              API参考
            </NuxtLink>
            <a href="https://github.com/wordflowlab/agentsdk" target="_blank" class="text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white">
              GitHub
            </a>
          </div>
        </div>
        <ColorModeToggle />
      </nav>
    </header>

    <div class="flex max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- 侧边栏 -->
      <aside class="hidden lg:block w-64 flex-shrink-0 pr-8">
        <div class="sticky top-24">
          <ContentNavigation v-slot="{ navigation }">
            <nav class="space-y-1">
              <template v-for="item in navigation" :key="item._path">
                <div class="mb-4">
                  <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">
                    {{ item.title }}
                  </h3>
                  <ul class="space-y-1">
                    <li v-for="child in item.children" :key="child._path">
                      <NuxtLink
                        :to="child._path"
                        class="block px-3 py-2 text-sm text-gray-700 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-300 dark:hover:text-white dark:hover:bg-gray-800 rounded-md"
                        active-class="nav-link-active bg-primary-50 dark:bg-primary-900"
                      >
                        {{ child.title }}
                      </NuxtLink>
                    </li>
                  </ul>
                </div>
              </template>
            </nav>
          </ContentNavigation>
        </div>
      </aside>

      <!-- 主内容区 -->
      <main class="flex-1 min-w-0">
        <article class="prose dark:prose-invert max-w-none">
          <slot />
        </article>
      </main>

      <!-- 右侧目录 -->
      <aside class="hidden xl:block w-64 flex-shrink-0 pl-8">
        <div class="sticky top-24">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">
            本页目录
          </h3>
          <nav class="text-sm space-y-2">
            <p class="text-gray-500 dark:text-gray-400">使用浏览器内置的页内跳转</p>
          </nav>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup>
</script>
