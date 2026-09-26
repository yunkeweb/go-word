import { defineConfig } from 'vitepress'

const github = 'https://github.com/yunkeweb/go-word'
const pkgGoDev = 'https://pkg.go.dev/github.com/yunkeweb/go-word'

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
    logo: undefined,
    socialLinks: [{ icon: 'github', link: github }],
    search: { provider: 'local' },
  },

  locales: {
    root: {
      label: 'English',
      lang: 'en',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/guide/getting-started' },
          { text: 'API', link: pkgGoDev },
          { text: 'GitHub', link: github },
        ],
        sidebar: {
          '/guide/': [
            {
              text: 'Guide',
              items: [
                { text: 'Quick Start', link: '/guide/getting-started' },
                { text: 'Basic DOM', link: '/guide/basics' },
                { text: 'Office Math (OMML)', link: '/guide/math' },
                { text: 'DrawingML & Charts', link: '/guide/drawing' },
                { text: 'Streaming Parser', link: '/guide/streaming' },
                { text: 'Template Engine v2', link: '/guide/template' },
                { text: 'Document Merger', link: '/guide/merger' },
                { text: 'Advanced Layout & Protect', link: '/guide/layout' },
              ],
            },
          ],
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
          { text: '指南', link: '/zh/guide/getting-started' },
          { text: 'API', link: pkgGoDev },
          { text: 'GitHub', link: github },
        ],
        sidebar: {
          '/zh/guide/': [
            {
              text: '指南',
              items: [
                { text: '快速开始', link: '/zh/guide/getting-started' },
                { text: '基础 DOM 操作', link: '/zh/guide/basics' },
                { text: 'Office Math 原生公式', link: '/zh/guide/math' },
                { text: 'DrawingML 图表与形状', link: '/zh/guide/drawing' },
                { text: '流式提取器', link: '/zh/guide/streaming' },
                { text: '模板引擎 v2', link: '/zh/guide/template' },
                { text: '文档无损合并', link: '/zh/guide/merger' },
                { text: '高级排版与保护', link: '/zh/guide/layout' },
              ],
            },
          ],
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
