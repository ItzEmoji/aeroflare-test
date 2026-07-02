import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    {
      type: 'category',
      label: 'Tutorials & Getting Started',
      items: [
        'tutorials/introduction',
        'tutorials/installation',
        'tutorials/quick-start',
        'tutorials/proxy-modes',
      ],
    },
    {
      type: 'category',
      label: 'How-to Guides',
      items: [
        'how-to/configuring-backends',
        'how-to/running-proxy',
        'how-to/cache-population',
        'how-to/cache-maintenance',
        'how-to/authentication',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      items: [
        'reference/cli',
        'reference/configuration',
        'reference/repository-layout',
      ],
    },
    {
      type: 'category',
      label: 'Concepts & Architecture',
      items: [
        'explanation/architecture',
        'explanation/oci-integration',
        'internals/architecture',
        'internals/subsystems',
        'internals/proxy-implementations',
        'internals/tasks-ui',
      ],
    },
  ],
};

export default sidebars;
