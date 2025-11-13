export default defineNuxtConfig({
  // 兼容性日期
  compatibilityDate: '2024-11-13',

  // 开发工具
  devtools: { enabled: true },

  // Nuxt模块
  modules: [
    '@nuxt/content',
    '@nuxtjs/tailwindcss',
    '@nuxtjs/color-mode'
  ],

  // 内容配置
  content: {
    documentDriven: true,
    highlight: {
      theme: 'github-dark',
      preload: ['go', 'bash', 'javascript', 'typescript', 'json', 'yaml', 'shell', 'sh']
    },
    markdown: {
      anchorLinks: true,
      remarkPlugins: [],
      rehypePlugins: []
    }
  },

  // 颜色模式
  colorMode: {
    classSuffix: '',
    preference: 'system',
    fallback: 'light'
  },

  // 应用配置
  app: {
    baseURL: '/agentsdk/',
    head: {
      title: 'AgentSDK - Go Agent Framework',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        {
          name: 'description',
          content: '企业级AI Agent运行时框架，事件驱动、云端沙箱、安全可控'
        },
        { property: 'og:title', content: 'AgentSDK Documentation' },
        { property: 'og:description', content: 'Go语言AI Agent开发框架完整文档' },
        { property: 'og:type', content: 'website' }
      ],
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/agentsdk/favicon.ico' }
      ]
    }
  },

  // CSS配置
  css: ['~/assets/css/main.css'],

  // 构建配置
  nitro: {
    prerender: {
      crawlLinks: true,
      routes: ['/'],
      failOnError: false
    }
  }
})
