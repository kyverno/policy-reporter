import theme from '@nuxt/content-theme-docs'
import highlightjs from 'highlight.js'
import path from 'path'

const baseURL = process.env.NODE_ENV === 'production' ? '/policy-reporter/' : ''

const config = theme({
  css: [
    'highlight.js/styles/vs2015.css',
    'assets/css/custom.css'
  ],
  docs: {
    primaryColor: '#E24F55'
  },
  content: {
    markdown: {
      highlighter(rawCode, lang) {
        const highlightedCode = highlightjs.highlight(rawCode, { language: lang }).value

        return `<pre class="highlight-pre-container"><code class="language-${lang} hljs">${highlightedCode}</code></pre>`
      }
    }
  },
  router: {
    base: baseURL
  },
  generate: {
    dir: '../docs'
  },
  hooks: {
    "vue-renderer:ssr:templateParams": function (params) {
      params.HEAD = params.HEAD.replace('<base href="/policy-reporter/">', "");
    }
  },
  content: {
    markdown: {
      rehypePlugins: [
        ['rehype-urls', (url) => {
          if (!baseURL) return;
          if (url.href && url.href.startsWith('/images/')) return path.join(baseURL, url.href);
        }]
      ]
    }
  },
  head: {
    link: [
      { rel: 'icon', type: 'image/x-icon', href: 'https://kyverno.github.io/policy-reporter/favicon.ico' }
    ]
  }
})

config.buildModules = config.buildModules.filter(module => module !== '@nuxtjs/pwa')

export default config
