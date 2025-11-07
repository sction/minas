import { RouteRecordRaw } from 'vue-router'

const depopsRoutes: RouteRecordRaw[] = [
  {
    name: 'depops_nodes',
    path: '/depops/nodes',
    component: () => import('@/pages/depops/NodeList.vue'),
    meta: {
      auth: 'depops.node.view',
    }
  },
  {
    name: 'depops_node_detail',
    path: '/depops/nodes/:id',
    component: () => import('@/pages/depops/NodeDetail.vue'),
    meta: {
      auth: 'depops.node.view',
    }
  },
  {
    name: 'depops_batch',
    path: '/depops/batch',
    component: () => import('@/pages/depops/BatchOperation.vue'),
    meta: {
      auth: 'depops.batch.execute',
    }
  },
]

export default depopsRoutes
