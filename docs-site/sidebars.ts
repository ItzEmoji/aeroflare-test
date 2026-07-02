import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    {
      type: 'category',
      label: 'Tutorials & Quick Starts',
      items: [
        'tutorials/quick-start',
        'tutorials/proxy-modes',
      ],
    },
    {
      type: 'category',
      label: 'CLI Reference',
      items: [
        'cli/core',
        'cli/auth',
        'cli/cache',
        'cli/maintenance',
      ],
    },
    {
      type: 'category',
      label: 'Architecture & Internals',
      items: [
        'internals/architecture',
        'internals/subsystems',
        'internals/proxy-implementations',
        'internals/tasks-ui',
      ],
    },
  ],
};

export default sidebars;
