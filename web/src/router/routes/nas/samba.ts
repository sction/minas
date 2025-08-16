export default [
    {
        name: "samba_list",
        path: "/nas/samba",
        component: () => import('@/pages/nas/samba/List.vue'),
        meta: {
            auth: 'samba.view',
        }
    },
    {
        name: "samba_new",
        path: "/nas/samba/new",
        component: () => import('@/pages/nas/samba/Edit.vue'),
        meta: {
            auth: 'samba.edit',
        }
    },
    {
        name: "samba_detail",
        path: "/nas/samba/:id",
        component: () => import('@/pages/nas/samba/View.vue'),
        meta: {
            auth: 'samba.view',
        }
    },
    {
        name: "samba_edit",
        path: "/nas/samba/:id/edit",
        component: () => import('@/pages/nas/samba/Edit.vue'),
        meta: {
            auth: 'samba.edit',
        }
    }
]