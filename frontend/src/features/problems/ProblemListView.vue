<script setup lang="ts">
import { computed, h, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  NButton,
  NDataTable,
  NSpace,
  NTag,
  useMessage,
  type DataTableColumns,
  type DataTableSortState,
} from 'naive-ui'
import { problemApi } from '@/api'
import type { ProblemListItem, ProblemQuery } from '@/api/types'
import PageHeader from '@/components/PageHeader.vue'
import MetaTag from '@/components/MetaTag.vue'
import ProblemFilters from './components/ProblemFilters.vue'
import { formatRelative } from '@/composables/format'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const loading = ref(false)
const rows = ref<ProblemListItem[]>([])
const total = ref(0)
const filtersExpanded = ref(false)

const filters = reactive<ProblemQuery>(readFiltersFromQuery())
const pagination = reactive({ page: 1, pageSize: 20 })
const sort = ref<string>((route.query.sort as string) || '-created_at')

function readFiltersFromQuery(): ProblemQuery {
  const q = route.query
  const arr = (v: unknown): string[] | undefined =>
    v === undefined ? undefined : Array.isArray(v) ? (v as string[]) : [v as string]
  return {
    q: (q.q as string) || undefined,
    project: arr(q.project),
    category_id: arr(q.category_id)?.map(Number),
    environment: arr(q.environment),
    severity: arr(q.severity),
    status: arr(q.status),
    tag: arr(q.tag),
  }
}

function currentQuery(): ProblemQuery {
  return {
    ...filters,
    sort: sort.value,
    page: pagination.page,
    limit: pagination.pageSize,
  }
}

async function fetchData() {
  loading.value = true
  try {
    const res = await problemApi.list(currentQuery())
    rows.value = res.items
    total.value = res.pagination.total
  } catch {
    message.error('Could not load problems')
  } finally {
    loading.value = false
  }
}

function syncUrl() {
  const query: Record<string, string | string[]> = {}
  for (const [k, v] of Object.entries(filters)) {
    if (v === undefined || v === '' || (Array.isArray(v) && !v.length)) continue
    query[k] = Array.isArray(v) ? v.map(String) : String(v)
  }
  if (sort.value && sort.value !== '-created_at') query.sort = sort.value
  if (pagination.page > 1) query.page = String(pagination.page)
  router.replace({ query })
}

function apply() {
  pagination.page = 1
  syncUrl()
  fetchData()
}

function reset() {
  filters.q = undefined
  filters.project = undefined
  filters.category_id = undefined
  filters.environment = undefined
  filters.severity = undefined
  filters.status = undefined
  filters.tag = undefined
  filters.created_from = filters.created_to = undefined
  filters.solved_from = filters.solved_to = undefined
  apply()
}

function onPageChange(page: number) {
  pagination.page = page
  syncUrl()
  fetchData()
}

function onSorterChange(s: DataTableSortState | DataTableSortState[] | null) {
  const state = Array.isArray(s) ? s[0] : s
  if (!state || !state.order) {
    sort.value = '-created_at'
  } else {
    const prefix = state.order === 'ascend' ? '' : '-'
    sort.value = `${prefix}${state.columnKey}`
  }
  apply()
}

const columns = computed<DataTableColumns<ProblemListItem>>(() => [
  {
    title: 'Title',
    key: 'title',
    sorter: true,
    render: (row) =>
      h(
        RouterLink,
        { to: { name: 'problem-detail', params: { id: row.id } }, class: 'row-title' },
        { default: () => row.title },
      ),
  },
  {
    title: 'Category',
    key: 'category',
    width: 130,
    render: (row) =>
      row.category
        ? h(NTag, { size: 'small', bordered: false }, { default: () => row.category!.name })
        : '—',
  },
  { title: 'Project', key: 'project', width: 150, ellipsis: { tooltip: true } },
  {
    title: 'Env',
    key: 'environment',
    width: 90,
    render: (row) => h(MetaTag, { kind: 'environment', value: row.environment }),
  },
  {
    title: 'Severity',
    key: 'severity',
    width: 110,
    sorter: true,
    render: (row) => h(MetaTag, { kind: 'severity', value: row.severity }),
  },
  {
    title: 'Status',
    key: 'status',
    width: 130,
    render: (row) => h(MetaTag, { kind: 'status', value: row.status }),
  },
  {
    title: 'Tags',
    key: 'tags',
    width: 200,
    render: (row) =>
      row.tags.length
        ? h(
            NSpace,
            { size: 4 },
            {
              default: () =>
                row.tags
                  .slice(0, 3)
                  .map((t) => h(NTag, { size: 'tiny', bordered: false }, { default: () => t.name })),
            },
          )
        : '—',
  },
  {
    title: 'Updated',
    key: 'updated_at',
    width: 120,
    sorter: true,
    render: (row) => h('span', { class: 'pkb-muted' }, formatRelative(row.updated_at)),
  },
])

watch(
  () => route.query,
  () => {
    Object.assign(filters, readFiltersFromQuery())
  },
)

onMounted(fetchData)
</script>

<template>
  <div>
    <PageHeader title="Problems" :subtitle="`${total} recorded`">
      <template #actions>
        <NButton size="small" @click="filtersExpanded = !filtersExpanded">
          {{ filtersExpanded ? 'Hide filters' : 'Filters' }}
        </NButton>
        <RouterLink :to="{ name: 'problem-new' }">
          <NButton size="small" type="primary">+ New Problem</NButton>
        </RouterLink>
      </template>
    </PageHeader>

    <ProblemFilters v-model="filters" :expanded="filtersExpanded" @apply="apply" @reset="reset" />

    <NDataTable
      remote
      :loading="loading"
      :columns="columns"
      :data="rows"
      :pagination="{
        page: pagination.page,
        pageSize: pagination.pageSize,
        itemCount: total,
        showQuickJumper: true,
      }"
      :bordered="false"
      :single-line="false"
      @update:page="onPageChange"
      @update:sorter="onSorterChange"
    />
  </div>
</template>

<style scoped>
:deep(.row-title) {
  font-weight: 550;
  color: #4f46e5;
}
:deep(.row-title:hover) {
  text-decoration: underline;
}
</style>
