import Upstreams from './pages/Upstreams';
import Schema from './pages/Schema';

export default [
    {
        component: Upstreams,
        exact: true,
        label: 'Upstream Servers',
        title: 'GraphQL Gateway | Upstream Servers',
        path: '/',
    },
    {
        component: Schema,
        exact: true,
        label: 'Gateway Schema',
        title: 'GraphQL Gateway | Schema',
        path: '/schema',
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