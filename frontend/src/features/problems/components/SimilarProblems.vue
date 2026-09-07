<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { NSpin } from 'naive-ui'
import { problemApi } from '@/api'
import type { ProblemListItem } from '@/api/types'
import MetaTag from '@/components/MetaTag.vue'

const props = defineProps<{ problemId: number }>()

const loading = ref(false)
const items = ref<ProblemListItem[]>([])

async function load() {
  loading.value = true
  try {
    items.value = await problemApi.similar(props.problemId, 6)
  } catch {
    items.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => props.problemId, load)
</script>

<template>
  <NSpin :show="loading">
    <ul v-if="items.length" class="sim">
      <li v-for="it in items" :key="it.id">
        <RouterLink :to="{ name: 'problem-detail', params: { id: it.id } }" class="sim__link">
          <span class="sim__title">{{ it.title }}</span>
          <span class="sim__meta">
            <MetaTag kind="status" :value="it.status" />
            <span class="pkb-muted">{{ it.project }}</span>
          </span>
        </RouterLink>
      </li>
    </ul>
    <p v-else class="pkb-muted empty">ยังไม่พบปัญหาที่คล้ายกัน</p>
  </NSpin>
</template>

<style scoped>
.sim {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.sim__link {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px;
  border-radius: 6px;
}
.sim__link:hover {
  background: #f4f5f7;
}
.sim__title {
  font-size: 13px;
  font-weight: 550;
  color: #4f46e5;
}
.sim__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}
.empty {
  font-size: 13px;
}
</style>
