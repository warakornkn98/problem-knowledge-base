<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NGrid, NGi, NSpin, NStatistic, useMessage } from 'naive-ui'
import { dashboardApi } from '@/api'
import type { CommonError, DashboardBucket, DashboardSummary } from '@/api/types'
import PageHeader from '@/components/PageHeader.vue'
import BarList from '@/components/BarList.vue'
import CopyButton from '@/components/CopyButton.vue'

const router = useRouter()
const message = useMessage()

const loading = ref(true)
const summary = ref<DashboardSummary | null>(null)
const categories = ref<DashboardBucket[]>([])
const projects = ref<DashboardBucket[]>([])
const commonErrors = ref<CommonError[]>([])

onMounted(async () => {
  try {
    const [s, c, p, e] = await Promise.all([
      dashboardApi.summary(),
      dashboardApi.categories(),
      dashboardApi.projects(),
      dashboardApi.commonErrors(),
    ])
    summary.value = s
    categories.value = c
    projects.value = p
    commonErrors.value = e
  } catch {
    message.error('Could not load the dashboard')
  } finally {
    loading.value = false
  }
})

function openCategory(item: { slug?: string }) {
  if (item.slug) router.push({ name: 'problems', query: { tab: 'category', category: item.slug } })
}
function openProject(item: { label: string }) {
  router.push({ name: 'problems', query: { project: item.label } })
}
function searchError(text: string) {
  router.push({ name: 'search', query: { q: text } })
}

const stats = () => [
  { label: 'Problems', value: summary.value?.total ?? 0 },
  { label: 'Open', value: summary.value?.open ?? 0 },
  { label: 'Investigating', value: summary.value?.investigating ?? 0 },
  { label: 'Solved', value: summary.value?.solved ?? 0 },
  { label: 'Known', value: summary.value?.known ?? 0 },
  { label: 'Today', value: summary.value?.today ?? 0 },
]
</script>

<template>
  <div>
    <PageHeader title="Dashboard" subtitle="ภาพรวมปัญหาของทีม และ Error ที่พบซ้ำบ่อย" />

    <NSpin :show="loading">
      <NGrid :cols="6" :x-gap="12" :y-gap="12" responsive="screen" item-responsive>
        <NGi v-for="s in stats()" :key="s.label" span="6 s:3 m:2 l:1">
          <NCard size="small">
            <NStatistic :label="s.label" :value="s.value" />
          </NCard>
        </NGi>
      </NGrid>

      <NGrid :cols="2" :x-gap="16" :y-gap="16" responsive="screen" item-responsive style="margin-top: 16px">
        <NGi span="2 m:1">
          <NCard title="Most Common Categories" size="small">
            <BarList :items="categories" clickable color="#4f46e5" @select="openCategory" />
          </NCard>
        </NGi>
        <NGi span="2 m:1">
          <NCard title="Projects with Most Problems" size="small">
            <BarList :items="projects" clickable color="#0891b2" @select="openProject" />
          </NCard>
        </NGi>
      </NGrid>

      <NCard title="Recurring Errors" size="small" style="margin-top: 16px">
        <div v-if="commonErrors.length" class="errs">
          <div v-for="err in commonErrors" :key="err.error_message" class="err">
            <code class="pkb-code err__msg" @click="searchError(err.error_message)">{{ err.error_message }}</code>
            <div class="err__meta">
              <span class="err__count">×{{ err.count }}</span>
              <CopyButton :text="err.error_message" size="tiny" label="" />
            </div>
          </div>
        </div>
        <p v-else class="pkb-muted">No repeated errors recorded yet.</p>
      </NCard>
    </NSpin>
  </div>
</template>

<style scoped>
.errs {
  display: flex;
  flex-direction: column;
}
.err {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid #eef0f3;
}
.err:last-child {
  border-bottom: none;
}
.err__msg {
  flex: 1;
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.err__msg:hover {
  color: #4f46e5;
}
.err__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.err__count {
  font-variant-numeric: tabular-nums;
  color: #6b7280;
  font-size: 13px;
}
</style>
