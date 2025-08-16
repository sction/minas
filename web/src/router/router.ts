import { nextTick } from 'vue'
import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { LoadingBarApi } from 'naive-ui'
import ForbiddenPage from '../pages/403.vue'
import NotFoundPage from '../pages/404.vue'
import LoginPage from '../pages/Login.vue'
import InitPage from '../pages/Init.vue'
import { store } from "../store";
import webdavRoute from "./routes/nas/webdav";
import nfsRoute from "./routes/nas/nfs";
import sambaRoute from "./routes/nas/samba";
import externalNasRoute from "./routes/nas/externalNas";
import projectdirRoute from "./routes/basic/projectdir";
import schTaskRoute from "./routes/sch/task";
import schLogRoute from "./routes/sch/log";
import termRoute from "./routes/term/index";
import sflowRoute from "./routes/sflow/index";
import sflowLogRoute from "./routes/sflow/log";
import { t } from "@/locales";
import { baseUrl, frontBaseUrl } from '@/config';

var loadingBar: LoadingBarApi;

export function initLoadingBar(bar: LoadingBarApi) {
  loadingBar = bar
}

export function go(name: string, params: any) {
  router.push({ name: name, params: params })
}

const routes: RouteRecordRaw[] = [
  {
    name: 'home',
    path: "/",
    component: () => import('../pages/Home.vue'),
    meta: {
      auth: '?',
    }
  },
  {
    name: 'login',
    path: '/login',
    component: LoginPage,
    meta: {
      layout: "empty",
      auth: '*',
    }
  },
  {
    name: 'init',
    path: '/init',
    component: InitPage,
    meta: {
      layout: "empty",
      auth: '*',
    }
  },
  {
    name: 'profile',
    path: "/profile",
    component: () => import('../pages/Profile.vue'),
    meta: {
      auth: '?',
    }
  },
  {
    name: "setting",
    path: "/system/settings",
    component: () => import('../pages/setting/Setting.vue'),
    meta: {
      auth: 'setting.view',
    }
  },
  {
    name: '403',
    path: '/403',
    component: ForbiddenPage,
    meta: {
      layout: "simple",
      auth: '*',
    }
  },
  {
    name: '404',
    path: '/404',
    component: NotFoundPage,
    meta: {
      layout: "simple",
      auth: '*',
    }
  },
  {
    name: 'not-found',
    path: '/:pathMatch(.*)*',
    redirect: { name: '404' }
  },
  ...webdavRoute,
  ...nfsRoute,
  ...sambaRoute,
  ...externalNasRoute,
  ...projectdirRoute,
  ...schTaskRoute,
  ...schLogRoute,
  ...termRoute,
  ...sflowRoute,
  ...sflowLogRoute,
]

function createSiteRouter() {
  const router = createRouter({
    history: createWebHistory(frontBaseUrl),
    routes,
  })
  router.beforeEach(function (to, from, next) {
    if (!from || to.path !== from.path) {
      loadingBar?.start()
      window.document.title = t(`titles.${to.name as string}`) + ' - Minas'
    }

    const auth = to.meta.auth || '*'
    if (auth !== '*') {
      if (store.getters.anonymous) {
        next({ name: 'login', query: { redirect: to.fullPath } })
        return
      }

      if (auth !== '?' && !store.getters.allow(auth)) {
        next({ name: '403' })
        return
      }
    }

    next()
  })

  router.afterEach(function (to, from) {
    if (!from || to.path !== from.path) {
      loadingBar?.finish()
      if (to.hash && to.hash !== from.hash) {
        nextTick(() => {
          const el = document.querySelector(to.hash)
          if (el) el.scrollIntoView()
        })
      }
    }
  })

  return router
}

export const router = createSiteRouter()