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
      items: ['funcionalidades/descoberta', 'funcionalidades/app-web', 'funcionalidades/alertas'],
    },
    {
      type: 'category',
      label: 'Técnico',
      collapsed: false,
      items: ['tecnico/arquitetura', 'tecnico/api', 'tecnico/implantacao', 'tecnico/modelo-de-ameacas', 'tecnico/seguranca-e-privacidade', 'tecnico/desenvolvimento'],
    },
  ],
};

export default sidebars;
