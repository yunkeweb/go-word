import { defineConfig } from 'vitepress'

const github = 'https://github.com/yunkeweb/go-word'
const pkgGoDev = 'https://pkg.go.dev/github.com/yunkeweb/go-word'

// Algolia DocSearch reservation. Leave appId/apiKey empty to keep local search.
// Set ALGOLIA_APP_ID / ALGOLIA_SEARCH_API_KEY / ALGOLIA_INDEX_NAME at build time
// (or fill the strings below) to switch the site to DocSearch.
const algolia = {
  appId: process.env.ALGOLIA_APP_ID ?? '',
  apiKey: process.env.ALGOLIA_SEARCH_API_KEY ?? '',
  indexName: process.env.ALGOLIA_INDEX_NAME ?? 'goword',
}

const enGuide = [
  { text: 'Quick Start & Installation', link: '/guide/getting-started' },
  { text: 'Core DOM', link: '/guide/basics' },
  { text: 'Office Math (OMML)', link: '/guide/math' },
  { text: 'DrawingML Charts & Shapes', link: '/guide/drawing' },
  { text: 'Streaming Parser', link: '/guide/streaming' },
  { text: 'Template Engine v2', link: '/guide/template' },
  { text: 'Document Merger', link: '/guide/merger' },
  { text: 'Advanced Layout & Security', link: '/guide/layout' },
  { text: 'Examples & Recipes', link: '/guide/examples' },
]

const zhGuide = [
  { text: '快速开始与安装', link: '/zh/guide/getting-started' },
  { text: '核心 DOM', link: '/zh/guide/basics' },
  { text: 'Office Math 原生公式', link: '/zh/guide/math' },
  { text: 'DrawingML 图表与形状', link: '/zh/guide/drawing' },
  { text: '流式解析器', link: '/zh/guide/streaming' },
  { text: '模板引擎 v2', link: '/zh/guide/template' },
  { text: '文档无损合并', link: '/zh/guide/merger' },
  { text: '高级排版与安全', link: '/zh/guide/layout' },
  { text: '实战案例库', link: '/zh/guide/examples' },
]

export default defineConfig({
  title: 'GoWord',
  description: 'Pure Go OpenXML Microsoft Word (.docx) Library',
  base: '/go-word/',
  lastUpdated: true,
  cleanUrls: true,
  ignoreDeadLinks: false,

  head: [
    ['meta', { name: 'theme-color', content: '#00ADD8' }],
  ],

  themeConfig: {
    socialLinks: [{ icon: 'github', link: github }],
    search: algolia.appId && algolia.apiKey
      ? { provider: 'algolia', options: algolia }
      : {
          provider: 'local',
          options: {
            locales: {
              zh: {
                translations: {
                  button: { buttonText: '搜索文档', buttonAriaLabel: '搜索文档' },
                  modal: {
                    noResultsText: '无法找到相关结果',
                    resetButtonTitle: '清除查询',
                    footer: { selectText: '选择', navigateText: '切换', closeText: '关闭' },
                  },
                },
              },
            },
          },
        },
  },

  locales: {
    root: {
      label: 'English',
      lang: 'en',
      themeConfig: {
        nav: [
          { text: 'Guide', items: enGuide },
          { text: 'Examples', link: '/guide/examples' },
          { text: 'API', link: pkgGoDev },
          { text: 'GitHub', link: github },
        ],
        sidebar: {
          '/guide/': [{ text: 'Guide', items: enGuide }],
        },
        outline: { label: 'On this page', level: [2, 3] },
        editLink: {
          pattern: `${github}/edit/main/docs/:path`,
          text: 'Edit this page on GitHub',
        },
        lastUpdated: { text: 'Updated' },
        docFooter: { prev: 'Previous', next: 'Next' },
        footer: {
          message: 'Released under the GNU LGPL v3.0.',
          copyright: 'Copyright © yunkeweb',
        },
      },
    },
    zh: {
      label: '简体中文',
      lang: 'zh-CN',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '指南', items: zhGuide },
          { text: '实战案例', link: '/zh/guide/examples' },
          { text: 'API', link: pkgGoDev },
          { text: 'GitHub', link: github },
        ],
        sidebar: {
          '/zh/guide/': [{ text: '指南', items: zhGuide }],
        },
        outline: { label: '本页目录', level: [2, 3] },
        editLink: {
          pattern: `${github}/edit/main/docs/:path`,
          text: '在 GitHub 上编辑此页',
        },
        lastUpdated: { text: '最后更新' },
        docFooter: { prev: '上一页', next: '下一页' },
        footer: {
          message: '基于 GNU LGPL v3.0 发布。',
          copyright: 'Copyright © yunkeweb',
        },
      },
    },
  },
})
