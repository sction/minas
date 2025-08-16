<template>
  <n-config-provider :theme="theme" :locale="locale" :date-locale="dateLocale" :theme-overrides="themeFixed"
    :hljs="hljs">
    <n-loading-bar-provider>
      <n-message-provider>
        <n-notification-provider>
          <n-dialog-provider>
            <root />
          </n-dialog-provider>
        </n-notification-provider>
      </n-message-provider>
    </n-loading-bar-provider>
    <n-global-style />
  </n-config-provider>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";
import {
  zhCN,
  dateZhCN,
  NConfigProvider,
  NDialogProvider,
  NNotificationProvider,
  NMessageProvider,
  NLoadingBarProvider,
  NGlobalStyle,
  useMessage,
  useLoadingBar,
  useDialog,
  useNotification,
} from "naive-ui";
import { darkTheme } from "naive-ui";
import { useRoute, useRouter } from "vue-router";
import { useStore } from "vuex";
import hljs from 'highlight.js/lib/core'
import json from 'highlight.js/lib/languages/json'
import yaml from 'highlight.js/lib/languages/yaml'
import bash from 'highlight.js/lib/languages/bash'
import { initLoadingBar } from "@/router/router";
import DefaultLayout from "./layouts/Default.vue";
import SimpleLayout from "./layouts/Simple.vue";
import EmptyLayout from "./layouts/Empty.vue";
import setIntroduction from '@/utils/setIconfont';
import { onBeforeMount } from "vue";
const Root = defineComponent({
  name: "App",
  components: {
    DefaultLayout,
    SimpleLayout,
    EmptyLayout,
  },
  template: '<component :is="layout"></component>',
  setup() {
    window.message = useMessage();
    window.router = useRouter();
    window.dialog = useDialog();
    window.notification = useNotification();
    initLoadingBar(useLoadingBar());

    // 设置初始化，防止刷新时恢复默认
    onBeforeMount(() => {
      // 设置批量第三方 icon 图标
      setIntroduction.cssCdn();
      // 设置批量第三方 js
      setIntroduction.jsCdn();
    });
    const route = useRoute();
    return {
      layout: computed(() => (route.meta.layout || "default") + "-layout"),
    };
  },
})

export default defineComponent({
  name: "App",
  components: {
    NConfigProvider,
    NDialogProvider,
    NNotificationProvider,
    NMessageProvider,
    NLoadingBarProvider,
    NGlobalStyle,
    Root,
  },
  setup() {
    const store = useStore();
    const theme = computed(() => store.state.preference?.theme === "dark" ? darkTheme : null);
    const locale = computed(() => store.state.preference?.locale === 'zh' ? zhCN : null);
    const dateLocale = computed(() => store.state.preference?.locale === 'zh' ? dateZhCN : null);
    const themeFixed = {
      Form: {
        feedbackHeightMedium: "20px",
        feedbackFontSizeMedium: "12px",
        // blankHeightMedium: "30px",
      },
    }

    hljs.registerLanguage('json', json)
    hljs.registerLanguage('yaml', yaml)
    hljs.registerLanguage('bash', bash)

    return {
      locale,
      dateLocale,
      theme,
      themeFixed,
      hljs,
    };
  },
});
</script>

<style>
@import '@/assets/common.scss';

#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>