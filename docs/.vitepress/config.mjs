import { defineConfig } from 'vitepress'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  title: 'WeLink',
  description: '微信聊天数据分析平台',
  lang: 'zh-CN',

  ignoreDeadLinks: [/^http:\/\/localhost/],

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
  ],

  themeConfig: {
    logo: '/logo.svg',
    siteTitle: 'WeLink',

    nav: [
      { text: '首页', link: '/' },
      { text: '下载安装', link: '/install' },
      { text: 'MCP Server', link: '/mcp-server' },
      { text: 'API 接口', link: '/api' },
      { text: '技术文档', link: '/README' },
    ],

    sidebar: [
      {
        text: '开始使用',
        items: [
          { text: '产品概览', link: '/' },
          { text: '下载与安装', link: '/install' },
          { text: 'MCP Server', link: '/mcp-server' },
        ],
      },
      {
        text: '技术参考',
        items: [
          { text: '文档总览', link: '/README' },
          { text: 'API 接口文档', link: '/api' },
          { text: '数据库结构', link: '/database' },
          { text: '索引与初始化', link: '/indexing' },
          { text: '情感分析', link: '/sentiment' },
          { text: '词云生成', link: '/wordcloud' },
        ],
      },
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/runzhliu/welink' },
    ],

    footer: {
      message: 'WeLink — 所有数据仅在本地处理，不上传任何服务器',
    },

    search: {
      provider: 'local',
    },
  },
})
