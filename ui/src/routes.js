import Upstreams from './pages/Upstreams';
import Schema from './pages/Schema';
import Console from './pages/Console';

export default [
  {
    component: Console,
    exact: true,
    label: 'GraphQL Console',
    title: 'GraphQL Gateway | Console',
    path: '/',
  },
  {
    component: Upstreams,
    exact: true,
    label: 'Upstream Servers',
    title: 'GraphQL Gateway | Upstream Servers',
    path: '/upstreams/',
  },
  {
    component: Schema,
    exact: false,
    label: 'Gateway Schema',
    title: 'GraphQL Gateway | Schema',
    path: '/schema/',
  },
  // {
  //   component: Support,
  //   exact: true,
  //   isAsync: true,
  //   label: 'Support',
  //   path: '/support',
  //   title: 'PatternFly Seed | Support Page',
  // },
];