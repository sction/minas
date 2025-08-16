import { h } from 'vue'
import { NIcon, MenuOption } from 'naive-ui'
import { RouteLocationNormalizedLoaded, RouterLink, useRoute, useRouter } from 'vue-router'
import { store } from "../store";
import {
  HomeOutline,
  PersonOutline,
  PeopleOutline,
  SettingsOutline,
  ConstructOutline,
  GridOutline,
  GlobeOutline,
  CubeOutline,
  BarChartOutline,
  LayersOutline,
  DocumentTextOutline,
  DocumentOutline,
  DocumentLockOutline,
  FileTrayFullOutline,
  BusinessOutline,
  ServerOutline,
  AlbumsOutline,
  ImageOutline,
  ImagesOutline,
  TerminalOutline,
  BookOutline,
  ArchiveOutline,
  FolderOpenOutline,
  CloudOutline,
  LinkOutline,
} from "@vicons/ionicons5";
import XIcon from "@/components/Icon.vue";
import { t } from "@/locales";
import { removePageData} from "@/utils/render";
import systemApi from '@/api/system';

const router = useRouter();
// 系统环境信息，用于判断操作系统类型
let systemEnvironment: any = null;

// 初始化系统环境信息
const initSystemEnvironment = async () => {
  if (!systemEnvironment) {
    try {
      const result = (await systemApi.getEnvironment()).data as any;
      if (result && result.success) {
        systemEnvironment = result.data;
      }
    } catch (error) {
      console.warn('Failed to get system environment:', error);
      // 如果获取失败，默认假设为Linux系统
      systemEnvironment = { os: 'linux' };
    }
  }
  return systemEnvironment;
};

// 检查当前系统是否支持指定功能
const isFeatureSupported = (feature: string): boolean => {
  if (!systemEnvironment) {
    // 如果环境信息还未加载，默认支持（后续会重新渲染）
    return true;
  }
  
  const os = systemEnvironment.os || 'linux';
  
  switch (feature) {
    case 'nfs':
    case 'samba':
      // NFS和Samba仅在Linux系统上支持
      return os === 'linux';
    default:
      return true;
  }
};

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: () => h(icon) });
}

export const renderMenuLabel = (option: any) => {
  removePageData(option.path);
  if (!('path' in option)) {
    return option.label
  }
  return h(
    RouterLink,
    {
      to: option.path
    },
    {
      default: () => option.label
    }
  )
}

export function findMenuValue(route: RouteLocationNormalizedLoaded): string {
  var path = route.path;
  do {
    const option = findOption(menuOptions, path)
    if (option) {
      return option.key
    } else {
      const index = path.lastIndexOf("/")
      if (index <= 0) {
        return ""
      }
      path = path.substring(0, index)
    }
  } while (true)
}

function findOption(options: MenuOption[], path: string): any {
  for (const option of options) {
    if (option.path === path) {
      return option
    } else if (option.children) {
      const opt = findOption(option.children, path)
      if (opt) return opt
    }
  }
  return null
}

export function findActiveOptions(route: RouteLocationNormalizedLoaded): MenuOption[] {
  const result: MenuOption[] = []
  findOptions(result, menuOptions, route.path)
  return result
}

function findOptions(result: MenuOption[], options: MenuOption[], path: string): boolean {
  for (const option of options) {
    if (option.path) {
      if (option.path != "/" && path.startsWith(<string>option.path)) {
        result.push(option)
        return true
      }
    } else if (option.children) {
      result.push(option)
      if (findOptions(result, option.children, path)) {
        return true
      } else {
        result.pop()
      }
    }
  }
  return false
}
export const allow = (menu: any) => {
  const auth = menu.meta?.auth || '*'
  if (auth !== '*') {
    if (store.getters.anonymous) {
      return false
    }

    if (auth !== '?' && !store.getters.allow(auth)) {
      return false
    }
  }
  return true;

}
export const getMenus = async () => {
  // 确保系统环境信息已加载
  await initSystemEnvironment();
  const ms = getMenuOptions(menuOptions);
  return ms;
}
const getMenuOptions = (menus: MenuOption[]) => {
  const newMenus: MenuOption[] = [];
  for (let i = 0; i < menus.length; i++) {
    const menu = menus[i] as any;
    if (menu.hide) {
      //隐藏菜单
      if (router && menu.path) {
        router.removeRoute(menu.path);
      }
      continue;
    }
    
    // 检查系统特性支持
    if (menu.requiresFeature && !isFeatureSupported(menu.requiresFeature)) {
      //系统不支持该特性，隐藏菜单
      if (router && menu.path) {
        router.removeRoute(menu.path);
      }
      continue;
    }
    
    if (!allow(menu)) {
      //无权限 隐藏菜单
      if (router && menu.path) {
        router.removeRoute(menu.path);
      }
      continue;
    }
    const newMenu: MenuOption = {
      label: menu.label,
      key: menu.key,
      icon: menu.icon,
    }
    if (menu.path) {
      newMenu.path = menu.path;
    }
    if (menu.children && menu.children.length > 0) {
      newMenu.children = getMenuOptions(menu.children);
    }
    newMenus.push(newMenu);
  }
  return newMenus;
}
export const menuOptions: MenuOption[] = [
  {
    label: t('fields.home'),
    key: "home",
    path: "/",
    icon: renderIcon(HomeOutline),
    meta: {
      auth: '?'
    }
  },
  {
    label: t('fields.nas'),
    key: "nas",
    icon: renderIcon(FileTrayFullOutline),
    children: [
      {
        label: t('nas.webdav'),
        key: "webdav",
        path: "/nas/webdav",
        icon: renderIcon(CloudOutline),
      },
      {
        label: t('nas.nfs'),
        key: "nfs",
        path: "/nas/nfs",
        icon: renderIcon(LayersOutline),
        requiresFeature: 'nfs', // 仅Linux系统支持
      },
      {
        label: t('nas.samba'),
        key: "samba",
        path: "/nas/samba",
        icon: renderIcon(FolderOpenOutline),
        requiresFeature: 'samba', // 仅Linux系统支持
      },
      {
        label: t('nas.external'),
        key: "external",
        path: "/nas/external",
        icon: renderIcon(LinkOutline),
      }
    ],
  },
  {
    label: t('fields.projectdir'),
    key: "projectdir",
    path: "/basic/projectdir",
    icon: renderIcon(AlbumsOutline),
  },
  // {
  //   label: '作业流程',
  //   key: "sflow",
  //   path: "/sflow",
  //   icon: renderIcon(GlobeOutline),
  // },
  {
    label: t('fields.schtask'),
    key: "schtask",
    path: "/schtask",
    icon: renderIcon(BookOutline),
  },
  {
    label: t('titles.term_index'),
    key: "term",
    path: "/term",
    icon: renderIcon(TerminalOutline),
  },
  {
    label: t('fields.system'),
    key: "system",
    hide: true,
    icon: renderIcon(SettingsOutline),
    children: [
      {
        label: t('objects.setting'),
        key: "config",
        path: "/system/settings",
        icon: renderIcon(ConstructOutline),
      },
    ],
  },
]
