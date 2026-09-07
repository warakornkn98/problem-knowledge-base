<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { NButton, NCheckbox, NEmpty, NInput, NPagination, NSpin, useMessage } from 'naive-ui'
import { searchApi } from '@/api'
import type { Facets, ProblemListItem, ProblemQuery } from '@/api/types'
import PageHeader from '@/components/PageHeader.vue'
import MetaTag from '@/components/MetaTag.vue'
import { useMetaStore } from '@/stores/meta'
import { renderHeadline } from '@/composables/highlight'
import { formatRelative } from '@/composables/format'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const meta = useMetaStore()

const q = ref((route.query.q as string) || '')
const loading = ref(false)
const items = ref<ProblemListItem[]>([])
const facets = ref<Facets | null>(null)
const total = ref(0)
const page = ref(1)
const limit = 20
const searched = ref(false)

const selected = reactive<Record<string, string[]>>({
  category: [],
  status: [],
  severity: [],
  environment: [],
  project: [],
  tag: [],
})

function buildQuery(): ProblemQuery {
  const categoryIds = selected.category
    .map((slug) => meta.categories.find((c) => c.slug === slug)?.id)
    .filter((v): v is number => typeof v === 'number')
  return {
    q: q.value || undefined,
    category_id: categoryIds.length ? categoryIds : undefined,
    tag: selected.tag,
    status: selected.status,
    severity: selected.severity,
    environment: selected.environment,
    project: selected.project,
    page: page.value,
    limit,
  }
}

async function run(resetPage = true) {
  if (resetPage) page.value = 1
  loading.value = true
  searched.value = true
  try {
    const res = await searchApi.search(buildQuery())
    items.value = res.items
    facets.value = res.facets
    total.value = res.pagination.total
    syncUrl()
  } catch {
    message.error('Search failed')
  } finally {
    loading.value = false
  }
}

function syncUrl() {
  router.replace({ query: { q: q.value || undefined } })
}

function toggle(group: keyof typeof selected, value: string) {
  const arr = selected[group]
  const i = arr.indexOf(value)
  if (i >= 0) arr.splice(i, 1)
  else arr.push(value)
  run()
}

function clearFacets() {
  for (const k of Object.keys(selected)) selected[k] = []
  run()
}

onMounted(() => {
  meta.load().catch(() => {})
  if (q.value) run()
})

watch(
  () => route.query.q,
  (v) => {
    if (typeof v === 'string' && v !== q.value) {
      q.value = v
      run()
    }
  },
)

const FACET_GROUPS: { key: keyof typeof selected; label: string; field: keyof Facets }[] = [
  { key: 'status', label: 'Status', field: 'statuses' },
  { key: 'severity', label: 'Severity', field: 'severities' },
  { key: 'environment', label: 'Environment', field: 'environments' },
  { key: 'category', label: 'Category', field: 'categories' },
  { key: 'project', label: 'Project', field: 'projects' },
  { key: 'tag', label: 'Tags', field: 'tags' },
]
</script>

<template>
  <div>
    <PageHeader
      title="Search"
      subtitle="ค้นหาจาก error, title, root cause, solution, prevention และ tags — เจอปัญหาเก่าที่ใกล้เคียงได้แม้จำชื่อไม่ได้"
    />

    <div class="bar">
      <NInput
        v-model:value="q"
        size="large"
        placeholder='ลองพิมพ์ เช่น  "connection timeout"'
        clearable
        @keyup.enter="run()"
      >
        <template #prefix>🔎</template>
      </NInput>
      <NButton size="large" type="primary" @click="run()">Search</NButton>
    </div>

    <div class="grid">
      <aside class="facets" v-if="searched">
        <div class="facets__head">
          <span>Filters</span>
          <NButton text size="tiny" @click="clearFacets">clear</NButton>
        </div>
        <template v-for="g in FACET_GROUPS" :key="g.key">
          <div v-if="facets && facets[g.field]?.length" class="facet">
            <div class="facet__label">{{ g.label }}</div>
            <div v-for="fc in facets[g.field]" :key="fc.value" class="facet__opt">
              <NCheckbox
                :checked="selected[g.key].includes(fc.value)"
                @update:checked="toggle(g.key, fc.value)"
              >
                <span class="facet__name">{{ fc.label }}</span>
                <span class="facet__count">{{ fc.count }}</span>
              </NCheckbox>
            </div>
          </div>
        </template>
      </aside>

      <section class="results">
        <NSpin :show="loading">
          <template v-if="searched">
            <p class="results__count pkb-muted">{{ total }} result{{ total === 1 ? '' : 's' }}</p>
            <ul v-if="items.length" class="hits">
              <li v-for="hit in items" :key="hit.id" class="hit">
                <RouterLink :to="{ name: 'problem-detail', params: { id: hit.id } }" class="hit__title">
                  {{ hit.title }}
                </RouterLink>
                <p v-if="hit.headline" class="hit__snippet pkb-code" v-html="renderHeadline(hit.headline)" />
                <div class="hit__meta">
                  <MetaTag kind="status" :value="hit.status" />
                  <MetaTag kind="severity" :value="hit.severity" />
                  <span class="pkb-muted">{{ hit.category?.name }} · {{ hit.project }}</span>
                  <span class="pkb-muted">· {{ formatRelative(hit.updated_at) }}</span>
                </div>
              </li>
            </ul>
            <NEmpty v-else description="ไม่พบปัญหาที่ตรงกับคำค้น" style="margin: 40px 0" />

            <NPagination
              v-if="total > limit"
              :page="page"
              :page-size="limit"
              :item-count="total"
              style="margin-top: 16px"
              @update:page="(p: number) => { page = p; run(false) }"
            />
          </template>
          <NEmpty v-else description="พิมพ์คำค้นแล้วกด Search" style="margin: 60px 0" />
        </NSpin>
      </section>
    </div>
  </div>
</template>

<style scoped>
.bar {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
}
.grid {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 20px;
}
@media (max-width: 860px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
.facets {
  background: #fff;
  border: 1px solid #e3e5e8;
  border-radius: 10px;
  padding: 14px;
  align-self: start;
}
.facets__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  font-size: 13px;
  margin-bottom: 10px;
}
.facet {
  margin-bottom: 14px;
}
.facet__label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #8a8f98;
  margin-bottom: 6px;
}
.facet__opt {
  padding: 2px 0;
}
.facet__name {
  font-size: 13px;
}
.facet__count {
  color: #9aa0a6;
  font-size: 12px;
  margin-left: 6px;
}
.hits {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.hit {
  background: #fff;
  border: 1px solid #e8eaed;
  border-radius: 8px;
  padding: 12px 14px;
}
.hit__title {
  font-weight: 600;
  color: #4f46e5;
}
.hit__title:hover {
  text-decoration: underline;
}
.hit__snippet {
  margin: 6px 0;
  color: #4b5563;
  background: #f7f8fa;
  border-radius: 6px;
  padding: 8px 10px;
  font-size: 12.5px;
}
.hit__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  font-size: 12px;
}
.results__count {
  font-size: 13px;
  margin: 0 0 10px;
}
</style>
