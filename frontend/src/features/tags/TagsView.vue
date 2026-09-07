<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  NButton,
  NDataTable,
  NInput,
  NInputGroup,
  NPopconfirm,
  NTag,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'
import { tagApi } from '@/api'
import type { Tag } from '@/api/types'
import { ApiError } from '@/api/client'
import { useMetaStore } from '@/stores/meta'
import PageHeader from '@/components/PageHeader.vue'

const router = useRouter()
const message = useMessage()
const meta = useMetaStore()

const loading = ref(false)
const rows = ref<Tag[]>([])
const filter = ref('')
const newTag = ref('')

const filtered = computed(() =>
  filter.value
    ? rows.value.filter((t) => t.name.includes(filter.value.toLowerCase()))
    : rows.value,
)

async function load() {
  loading.value = true
  try {
    rows.value = await tagApi.list()
  } catch {
    message.error('Could not load tags')
  } finally {
    loading.value = false
  }
}

async function create() {
  const name = newTag.value.trim().toLowerCase()
  if (!name) return
  try {
    await tagApi.create(name)
    newTag.value = ''
    await load()
    await meta.refreshTags()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Could not create tag')
  }
}

async function remove(row: Tag) {
  try {
    await tagApi.remove(row.id)
    await load()
    await meta.refreshTags()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Delete failed')
  }
}

const columns: DataTableColumns<Tag> = [
  {
    title: 'Tag',
    key: 'name',
    render: (row) =>
      h(NTag, { bordered: false, round: true }, { default: () => row.name }),
  },
  {
    title: 'Used by',
    key: 'usage_count',
    width: 120,
    render: (row) =>
      h(
        NButton,
        {
          text: true,
          type: 'primary',
          disabled: row.usage_count === 0,
          onClick: () => router.push({ name: 'problems', query: { tag: row.slug } }),
        },
        { default: () => `${row.usage_count} problem${row.usage_count === 1 ? '' : 's'}` },
      ),
  },
  {
    title: '',
    key: 'actions',
    width: 90,
    render: (row) =>
      h(
        NPopconfirm,
        { onPositiveClick: () => remove(row) },
        {
          trigger: () =>
            h(NButton, { size: 'tiny', quaternary: true, type: 'error' }, { default: () => 'Delete' }),
          default: () =>
            row.usage_count > 0
              ? `This tag is on ${row.usage_count} problem(s). Remove anyway?`
              : 'Delete this tag?',
        },
      ),
  },
]

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="Tags" subtitle="ป้ายกำกับปัญหา ค้นหาและกรองได้เร็วขึ้น" />

    <div class="toolbar">
      <NInput v-model:value="filter" placeholder="Filter tags…" clearable style="max-width: 220px" />
      <NInputGroup style="max-width: 280px">
        <NInput v-model:value="newTag" placeholder="new-tag" @keyup.enter="create" />
        <NButton type="primary" @click="create">Add</NButton>
      </NInputGroup>
    </div>

    <NDataTable :loading="loading" :columns="columns" :data="filtered" :bordered="false" />
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  justify-content: space-between;
}
</style>
