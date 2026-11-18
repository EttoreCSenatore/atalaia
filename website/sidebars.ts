import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
    'licenca',
    'roadmap',
    {
      type: 'category',
      label: 'Funcionalidades',
      collapsed: false,
      items: ['funcionalidades/descoberta'],
    },
    {
      type: 'category',
      label: 'Técnico',
      collapsed: false,
      items: ['tecnico/arquitetura', 'tecnico/modelo-de-ameacas', 'tecnico/seguranca-e-privacidade', 'tecnico/desenvolvimento'],
    },
  ],
};

export default sidebars;
