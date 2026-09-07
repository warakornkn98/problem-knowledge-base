<script setup lang="ts">
import { computed } from 'vue'
import CopyButton from './CopyButton.vue'

const props = withDefaults(
  defineProps<{ content: string; title?: string; emptyText?: string }>(),
  { title: '', emptyText: '—' },
)

const isEmpty = computed(() => !props.content || !props.content.trim())
</script>

<template>
  <div class="cb">
    <div class="cb__bar">
      <span class="cb__title">{{ title || 'output' }}</span>
      <CopyButton v-if="!isEmpty" :text="content" size="tiny" />
    </div>
    <pre v-if="!isEmpty" class="pkb-code cb__body">{{ content }}</pre>
    <div v-else class="cb__empty pkb-muted">{{ emptyText }}</div>
  </div>
</template>

<style scoped>
.cb {
  border: 1px solid #e3e5e8;
  border-radius: 8px;
  overflow: hidden;
  background: #0f172a;
}
.cb__bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 8px 4px 12px;
  background: #1e293b;
  border-bottom: 1px solid #334155;
}
.cb__title {
  font-size: 11px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: #94a3b8;
}
.cb__body {
  margin: 0;
  padding: 12px;
  color: #e2e8f0;
  overflow-x: auto;
}
.cb__empty {
  padding: 12px;
  background: #fff;
}
</style>
