import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
    'licenca',
    'roadmap',
    {
      type: 'category',
      label: 'Manual',
      collapsed: false,
      items: ['manual/primeiros-passos', 'manual/tarefas', 'manual/faq', 'manual/glossario'],
    },
    {
      type: 'category',
      label: 'Funcionalidades',
      collapsed: false,
      items: ['funcionalidades/descoberta', 'funcionalidades/app-web', 'funcionalidades/alertas', 'funcionalidades/seguranca-passiva', 'funcionalidades/ferramentas'],
    },
    {
      type: 'category',
      label: 'Operação',
      collapsed: false,
      items: ['operacao/configuracao', 'operacao/backup-e-restauracao', 'operacao/atualizacao', 'tecnico/monitoramento'],
    },
    {
      type: 'category',
      label: 'Técnico',
      collapsed: false,
      items: ['tecnico/arquitetura', 'tecnico/api', 'tecnico/implantacao', 'tecnico/modelo-de-ameacas', 'tecnico/seguranca-e-privacidade', 'tecnico/desenvolvimento', 'tecnico/versionamento', 'tecnico/cadeia-de-suprimentos', 'tecnico/decisoes', 'tecnico/comparacao-kaladrius'],
    },
  ],
};

export default sidebars;
