export default [
    {
        name: "nfs_list",
        path: "/nas/nfs",
        component: () => import('@/pages/nas/nfs/List.vue'),
        meta: {
            auth: 'nfs.view',
        }
    },
    {
        name: "nfs_new",
        path: "/nas/nfs/new",
        component: () => import('@/pages/nas/nfs/Edit.vue'),
        meta: {
            auth: 'nfs.edit',
        }
    },
    {
        name: "nfs_detail",
        path: "/nas/nfs/:id",
        component: () => import('@/pages/nas/nfs/View.vue'),
        meta: {
            auth: 'nfs.view',
        }
    },
    {
        name: "nfs_edit",
        path: "/nas/nfs/:id/edit",
        component: () => import('@/pages/nas/nfs/Edit.vue'),
        meta: {
            auth: 'nfs.edit',
        }
    }
]