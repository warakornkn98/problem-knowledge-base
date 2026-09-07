<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import {
  NButton,
  NDropdown,
  NIcon,
  NLayout,
  NLayoutSider,
  NMenu,
  NScrollbar,
  type MenuOption,
} from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { useMetaStore } from '@/stores/meta'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const meta = useMetaStore()

const collapsed = ref(false)

onMounted(() => {
  meta.load().catch(() => {})
})

function icon(path: string) {
  return () => h(NIcon, null, { default: () => h('svg', { viewBox: '0 0 24 24', width: '1em', height: '1em' }, [h('path', { fill: 'currentColor', d: path })]) })
}

const menuOptions = computed<MenuOption[]>(() => [
  {
    label: () => h(RouterLink, { to: { name: 'dashboard' } }, { default: () => 'Dashboard' }),
    key: 'dashboard',
    icon: icon('M13 3v6h8V3h-8M3 13v8h8v-8H3M3 3v8h8V3H3m10 10v8h8v-8h-8Z'),
  },
  {
    label: () => h(RouterLink, { to: { name: 'problems' } }, { default: () => 'Problems' }),
    key: 'problems',
    icon: icon('M12 2L1 21h22L12 2m0 3.99L19.53 19H4.47L12 5.99M11 10v4h2v-4h-2m0 6v2h2v-2h-2Z'),
  },
  {
    label: () => h(RouterLink, { to: { name: 'search' } }, { default: () => 'Search' }),
    key: 'search',
    icon: icon('M9.5 3A6.5 6.5 0 0 1 16 9.5c0 1.61-.59 3.09-1.56 4.23l.27.27h.79l5 5l-1.5 1.5l-5-5v-.79l-.27-.27A6.516 6.516 0 0 1 9.5 16A6.5 6.5 0 0 1 3 9.5A6.5 6.5 0 0 1 9.5 3m0 2C7 5 5 7 5 9.5S7 14 9.5 14S14 12 14 9.5S12 5 9.5 5Z'),
  },
  {
    label: () => h(RouterLink, { to: { name: 'categories' } }, { default: () => 'Categories' }),
    key: 'categories',
    icon: icon('M12 2l-5.5 9h11L12 2M17.5 17.5m-4.5 0a4.5 4.5 0 1 0 9 0a4.5 4.5 0 1 0-9 0M3 13.5h8v8H3v-8Z'),
  },
  {
    label: () => h(RouterLink, { to: { name: 'tags' } }, { default: () => 'Tags' }),
    key: 'tags',
    icon: icon('M5.5 7A1.5 1.5 0 0 1 4 5.5A1.5 1.5 0 0 1 5.5 4A1.5 1.5 0 0 1 7 5.5A1.5 1.5 0 0 1 5.5 7m15.91 4.58l-9-9C12.05 2.22 11.55 2 11 2H4c-1.11 0-2 .89-2 2v7c0 .55.22 1.05.59 1.41l9 9c.36.36.86.59 1.41.59c.55 0 1.05-.23 1.41-.59l7-7c.36-.36.59-.86.59-1.41c0-.55-.23-1.06-.59-1.42Z'),
  },
])

const activeKey = computed(() => {
  if (route.name === 'problem-detail' || route.name === 'problem-new' || route.name === 'problem-edit')
    return 'problems'
  return (route.name as string) ?? 'dashboard'
})

const userMenu = [
  { label: 'Sign out', key: 'logout' },
]

function onUserSelect(key: string) {
  if (key === 'logout') {
    auth.clear()
    router.push({ name: 'login' })
  }
}

const initials = computed(() => {
  const n = auth.user?.display_name || auth.user?.username || '?'
  return n.slice(0, 2).toUpperCase()
})
</script>

<template>
  <NLayout has-sider style="height: 100vh">
    <NLayoutSider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="230"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <div class="brand" :class="{ 'brand--mini': collapsed }">
        <span class="brand__mark">PKB</span>
        <span v-if="!collapsed" class="brand__name">Knowledge Base</span>
      </div>
      <NMenu
        :value="activeKey"
        :collapsed="collapsed"
        :collapsed-width="64"
        :options="menuOptions"
        :root-indent="18"
      />
    </NLayoutSider>

    <NLayout>
      <header class="topbar">
        <RouterLink :to="{ name: 'problem-new' }">
          <NButton type="primary" size="small">+ New Problem</NButton>
        </RouterLink>
        <div class="spacer" />
        <NDropdown :options="userMenu" @select="onUserSelect" trigger="click">
          <button class="user">
            <span class="user__avatar">{{ initials }}</span>
            <span class="user__name">{{ auth.user?.display_name || auth.user?.username }}</span>
          </button>
        </NDropdown>
      </header>

      <NScrollbar style="max-height: calc(100vh - 52px)">
        <main class="content">
          <RouterView />
        </main>
      </NScrollbar>
    </NLayout>
  </NLayout>
</template>

<style scoped>
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 18px;
  height: 52px;
}
.brand--mini {
  justify-content: center;
  padding: 16px 0;
}
.brand__mark {
  font-weight: 800;
  letter-spacing: 0.04em;
  background: #4f46e5;
  color: #fff;
  border-radius: 6px;
  padding: 3px 6px;
  font-size: 13px;
}
.brand__name {
  font-weight: 650;
  font-size: 14px;
  white-space: nowrap;
}
.topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 52px;
  padding: 0 20px;
  background: #fff;
  border-bottom: 1px solid #e3e5e8;
}
.spacer {
  flex: 1;
}
.user {
  display: flex;
  align-items: center;
  gap: 8px;
  border: none;
  background: transparent;
  cursor: pointer;
  padding: 4px 6px;
  border-radius: 6px;
  font: inherit;
}
.user:hover {
  background: #f1f2f4;
}
.user__avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: #4f46e5;
  color: #fff;
  font-size: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}
.user__name {
  font-size: 13px;
  font-weight: 550;
}
.content {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}
</style>
