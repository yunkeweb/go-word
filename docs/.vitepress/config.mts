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

function enSidebar() {
  return [
    {
      text: 'Getting Started',
      collapsed: false,
      items: [
        { text: 'Installation', link: '/guide/installation' },
        { text: 'Quick Start', link: '/guide/getting-started' },
        { text: 'Architecture', link: '/guide/architecture' },
        { text: 'OpenXML Compatibility', link: '/guide/compatibility' },
      ],
    },
    {
      text: 'Basic Elements',
      collapsed: false,
      items: [
        { text: 'Paragraphs & Runs', link: '/guide/paragraph' },
        { text: 'Tables & Nested Cells', link: '/guide/table' },
        { text: 'Images & Shapes', link: '/guide/image' },
        { text: 'Headers, Footers & Page Numbers', link: '/guide/header-footer' },
        { text: 'Watermark & Document Protection', link: '/guide/protect' },
      ],
    },
    {
      text: 'Advanced Layout & Office Math',
      collapsed: false,
      items: [
        { text: 'Multi-Column Layout (w:cols)', link: '/guide/columns' },
        { text: 'Table of Contents', link: '/guide/toc' },
        { text: 'LaTeX → OMML (AddMath)', link: '/guide/math' },
      ],
    },
    {
      text: 'DrawingML Visuals',
      collapsed: false,
      items: [
        { text: 'Charts (Bar / Line / Pie / Area / Combo)', link: '/guide/charts' },
        { text: 'Shapes & Text Boxes (wps:wsp)', link: '/guide/shapes' },
      ],
    },
    {
      text: 'Enterprise Features',
      collapsed: false,
      items: [
        { text: 'O(1) Streaming Extractor', link: '/guide/streaming' },
        { text: 'Template Engine v2 (Pipes)', link: '/guide/template' },
        { text: 'Lossless Document Merger', link: '/guide/merger' },
      ],
    },
    {
      text: 'Recipes',
      collapsed: false,
      items: [
        { text: 'Enterprise Recipes', link: '/guide/recipes' },
      ],
    },
    {
      text: 'Benchmarks & FAQ',
      collapsed: false,
      items: [
        { text: 'Benchmarks', link: '/guide/benchmarks' },
        { text: 'FAQ', link: '/guide/faq' },
      ],
    },
  ]
}

function zhSidebar() {
  return [
    {
      text: '指南',
      collapsed: false,
      items: [
        { text: '安装', link: '/zh/guide/installation' },
        { text: '快速开始', link: '/zh/guide/getting-started' },
        { text: '架构设计', link: '/zh/guide/architecture' },
        { text: 'OpenXML 兼容性', link: '/zh/guide/compatibility' },
      ],
    },
    {
      text: '基础元素',
      collapsed: false,
      items: [
        { text: '段落与 Run', link: '/zh/guide/paragraph' },
        { text: '表格与嵌套单元格', link: '/zh/guide/table' },
        { text: '图片与 Shape', link: '/zh/guide/image' },
        { text: '页眉页脚与页码', link: '/zh/guide/header-footer' },
        { text: '水印与只读保护', link: '/zh/guide/protect' },
      ],
    },
    {
      text: '高级排版与数学',
      collapsed: false,
      items: [
        { text: '多栏排版 (w:cols)', link: '/zh/guide/columns' },
        { text: 'TOC 自动目录', link: '/zh/guide/toc' },
        { text: 'LaTeX 转 OMML 原生公式', link: '/zh/guide/math' },
      ],
    },
    {
      text: '图表与形状',
      collapsed: false,
      items: [
        { text: '柱 / 条 / 折 / 饼 / 面积 / 双轴组合图', link: '/zh/guide/charts' },
        { text: 'DrawingML 矢量形状与文本框', link: '/zh/guide/shapes' },
      ],
    },
    {
      text: '企业级应用',
      collapsed: false,
      items: [
        { text: 'O(1) 流式提取器', link: '/zh/guide/streaming' },
        { text: '模板 Pipe 过滤器', link: '/zh/guide/template' },
        { text: '多文档无损合并', link: '/zh/guide/merger' },
      ],
    },
    {
      text: '实战案例',
      collapsed: false,
      items: [
        { text: '企业级 Recipes', link: '/zh/guide/recipes' },
      ],
    },
    {
      text: '性能基准与常见问题',
      collapsed: false,
      items: [
        { text: '性能基准', link: '/zh/guide/benchmarks' },
        { text: '常见问题', link: '/zh/guide/faq' },
      ],
    },
  ]
}

export default defineConfig({
  title: 'GoWord',
  description: '100% pure Go standard-library engine for industrial Microsoft Word (.docx) processing, with zero third-party dependencies.',
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
          {
            text: 'Guide',
            items: [
              { text: 'Installation', link: '/guide/installation' },
              { text: 'Quick Start', link: '/guide/getting-started' },
              { text: 'Architecture', link: '/guide/architecture' },
              { text: 'OpenXML Compatibility', link: '/guide/compatibility' },
            ],
          },
          {
            text: 'Elements',
            items: [
              { text: 'Paragraphs & Runs', link: '/guide/paragraph' },
              { text: 'Tables', link: '/guide/table' },
              { text: 'Images & Shapes', link: '/guide/image' },
              { text: 'Headers & Footers', link: '/guide/header-footer' },
              { text: 'Protection', link: '/guide/protect' },
            ],
          },
          {
            text: 'Advanced',
            items: [
              { text: 'Columns', link: '/guide/columns' },
              { text: 'TOC', link: '/guide/toc' },
              { text: 'Office Math', link: '/guide/math' },
              { text: 'Charts', link: '/guide/charts' },
              { text: 'DrawingML Shapes', link: '/guide/shapes' },
            ],
          },
          {
            text: 'Enterprise',
            items: [
              { text: 'Streaming Parser', link: '/guide/streaming' },
              { text: 'Template Engine v2', link: '/guide/template' },
              { text: 'Document Merger', link: '/guide/merger' },
              { text: 'Recipes', link: '/guide/recipes' },
              { text: 'Benchmarks', link: '/guide/benchmarks' },
              { text: 'FAQ', link: '/guide/faq' },
            ],
          },
          { text: 'API', link: pkgGoDev },
          { text: 'GitHub', link: github },
        ],
        sidebar: {
          '/guide/': enSidebar(),
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
          {
            text: '指南',
            items: [
              { text: '安装', link: '/zh/guide/installation' },
              { text: '快速开始', link: '/zh/guide/getting-started' },
              { text: '架构设计', link: '/zh/guide/architecture' },
              { text: 'OpenXML 兼容性', link: '/zh/guide/compatibility' },
            ],
          },
          {
            text: '基础元素',
            items: [
              { text: '段落与 Run', link: '/zh/guide/paragraph' },
              { text: '表格', link: '/zh/guide/table' },
              { text: '图片与 Shape', link: '/zh/guide/image' },
              { text: '页眉页脚', link: '/zh/guide/header-footer' },
              { text: '水印与保护', link: '/zh/guide/protect' },
            ],
          },
          {
            text: '高级',
            items: [
              { text: '多栏排版', link: '/zh/guide/columns' },
              { text: 'TOC', link: '/zh/guide/toc' },
              { text: 'Office Math', link: '/zh/guide/math' },
              { text: '图表', link: '/zh/guide/charts' },
              { text: 'DrawingML 形状', link: '/zh/guide/shapes' },
            ],
          },
          {
            text: '企业级',
            items: [
              { text: '流式解析', link: '/zh/guide/streaming' },
              { text: '模板引擎 v2', link: '/zh/guide/template' },
              { text: '文档合并', link: '/zh/guide/merger' },
              { text: '实战案例', link: '/zh/guide/recipes' },
              { text: '性能基准', link: '/zh/guide/benchmarks' },
              { text: '常见问题', link: '/zh/guide/faq' },
            ],
          },
          { text: 'API', link: pkgGoDev },
          { text: 'GitHub', link: github },
        ],
        sidebar: {
          '/zh/guide/': zhSidebar(),
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
