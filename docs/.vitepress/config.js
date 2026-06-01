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
            { text: 'Device Power', link: '/metrics/device-power' },
            { text: 'Directory Info', link: '/metrics/directory-info' },
            { text: 'History', link: '/metrics/history' },
            { text: 'Job Queue', link: '/metrics/job-queue' },
            { text: 'MMU', link: '/metrics/mmu' },
            { text: 'Network Stats', link: '/metrics/network-stats' },
            { text: 'Printer Objects', link: '/metrics/printer-objects' },
            { text: 'Process Stats', link: '/metrics/process-stats' },
            { text: 'Server Info', link: '/metrics/server-info' },
            { text: 'Spoolman', link: '/metrics/spoolman' },
            { text: 'System Info', link: '/metrics/system-info' },
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
