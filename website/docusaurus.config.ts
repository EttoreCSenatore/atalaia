import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Atalaia',
  tagline: 'Vigia da sua rede: descubra dispositivos, acompanhe o histórico e receba alertas.',
  favicon: 'img/favicon.svg',

  future: {
    v4: true,
  },

  url: 'https://ettorecsenatore.github.io',
  baseUrl: '/atalaia/',
  organizationName: 'EttoreCSenatore',
  projectName: 'atalaia',
  trailingSlash: false,

  onBrokenLinks: 'throw',

  markdown: {
    mermaid: true,
    hooks: {
      onBrokenMarkdownLinks: 'throw',
    },
  },
  themes: ['@docusaurus/theme-mermaid'],

  i18n: {
    defaultLocale: 'pt-BR',
    locales: ['pt-BR', 'en'],
    localeConfigs: {
      'pt-BR': {label: 'Português (BR)', htmlLang: 'pt-BR'},
      en: {label: 'English', htmlLang: 'en'},
    },
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          showLastUpdateTime: false,
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    colorMode: {
      defaultMode: 'light',
      respectPrefersColorScheme: true,
    },
    mermaid: {
      theme: {light: 'neutral', dark: 'dark'},
    },
    navbar: {
      title: 'Atalaia',
      logo: {
        alt: 'Logo do Atalaia',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Documentação',
        },
        {to: '/docs/roadmap', label: 'Roadmap', position: 'left'},
        {type: 'localeDropdown', position: 'right'},
        {
          href: 'https://github.com/EttoreCSenatore/atalaia',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Produto',
          items: [
            {label: 'Visão geral', to: '/docs/intro'},
            {label: 'Roadmap', to: '/docs/roadmap'},
          ],
        },
        {
          title: 'Funcionalidades',
          items: [
            {label: 'Descoberta de dispositivos', to: '/docs/funcionalidades/descoberta'},
            {label: 'App web', to: '/docs/funcionalidades/app-web'},
          ],
        },
        {
          title: 'Técnico',
          items: [
            {label: 'Arquitetura', to: '/docs/tecnico/arquitetura'},
            {label: 'API e servidor', to: '/docs/tecnico/api'},
            {label: 'Implantação', to: '/docs/tecnico/implantacao'},
            {label: 'Modelo de ameaças', to: '/docs/tecnico/modelo-de-ameacas'},
            {label: 'Segurança e privacidade', to: '/docs/tecnico/seguranca-e-privacidade'},
            {label: 'Desenvolvimento', to: '/docs/tecnico/desenvolvimento'},
          ],
        },
        {
          title: 'Projeto',
          items: [
            {label: 'Licença', to: '/docs/licenca'},
            {label: 'GitHub', href: 'https://github.com/EttoreCSenatore/atalaia'},
          ],
        },
      ],
      copyright: `Atalaia © ${new Date().getFullYear()} Ettore Senatore. Software livre (AGPL-3.0-or-later). Feito com Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['json', 'bash', 'go', 'sql'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
