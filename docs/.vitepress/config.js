import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

export default withMermaid(defineConfig({
  title: 'Prometheus Klipper Exporter',
  description: 'Prometheus exporter for Klipper 3D printer firmware',
  base: '/prometheus-klipper-exporter/',

  themeConfig: {
    nav: [
      { text: 'Guide', link: '/guide/' },
      { text: 'Metrics', link: '/metrics/' },
      { text: 'Developers', link: '/developers/' },
      { text: 'GitHub', link: 'https://github.com/scross01/prometheus-klipper-exporter' },
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Configuration', link: '/guide/configuration' },
            { text: 'Authentication', link: '/guide/authentication' },
          ],
        },
      ],
      '/metrics/': [
        {
          text: 'Metrics Reference',
          items: [
            { text: 'Summary', link: '/metrics/' },
            { text: 'Server Info', link: '/metrics/server-info' },
            { text: 'Process Stats', link: '/metrics/process-stats' },
            { text: 'Network Stats', link: '/metrics/network-stats' },
            { text: 'Power Devices', link: '/metrics/device-power' },
            { text: 'System Info', link: '/metrics/system-info' },
            { text: 'Job Queue', link: '/metrics/job-queue' },
            { text: 'Directory Info', link: '/metrics/directory-info' },
            { text: 'History', link: '/metrics/history' },
            { text: 'Printer Objects', link: '/metrics/printer-objects' },
            { text: 'MMU', link: '/metrics/mmu' },
          ],
        },
      ],
      '/developers/': [
        {
          text: 'Developers',
          items: [
            { text: 'Overview', link: '/developers/' },
          ],
        },
      ],
    },

    search: {
      provider: 'local',
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/scross01/prometheus-klipper-exporter' },
    ],

    footer: {
      message: 'Released under the MIT License.',
    },
  },
}))
