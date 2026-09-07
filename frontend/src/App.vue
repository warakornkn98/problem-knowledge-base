<script setup lang="ts">
import { NConfigProvider, NDialogProvider, NLoadingBarProvider, NMessageProvider, NNotificationProvider } from 'naive-ui'
import type { GlobalThemeOverrides } from 'naive-ui'
import { RouterView } from 'vue-router'

// A restrained, developer-tool palette: indigo primary, tighter radii.
const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#4f46e5',
    primaryColorHover: '#4338ca',
    primaryColorPressed: '#3730a3',
    primaryColorSuppl: '#4338ca',
    borderRadius: '6px',
    fontSize: '14px',
  },
}
</script>

<template>
  <NConfigProvider :theme-overrides="themeOverrides">
    <NLoadingBarProvider>
      <NMessageProvider>
        <NNotificationProvider>
          <NDialogProvider>
            <RouterView />
          </NDialogProvider>
        </NNotificationProvider>
      </NMessageProvider>
    </NLoadingBarProvider>
  </NConfigProvider>
</template>
