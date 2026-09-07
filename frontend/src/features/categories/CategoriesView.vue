<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import {
  NButton,
  NColorPicker,
  NDataTable,
  NInput,
  NModal,
  NPopconfirm,
  NTag,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'
import { categoryApi } from '@/api'
import type { Category } from '@/api/types'
import { ApiError } from '@/api/client'
import { useMetaStore } from '@/stores/meta'
import PageHeader from '@/components/PageHeader.vue'

const message = useMessage()
const meta = useMetaStore()

const loading = ref(false)
const rows = ref<Category[]>([])
const showModal = ref(false)
const editing = ref<Category | null>(null)
const form = reactive({ name: '', description: '', color: '#4f46e5' })

async function load() {
  loading.value = true
  try {
    rows.value = await categoryApi.list()
  } catch {
    message.error('Could not load categories')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, { name: '', description: '', color: '#4f46e5' })
  showModal.value = true
}
function openEdit(row: Category) {
  editing.value = row
  Object.assign(form, { name: row.name, description: row.description, color: row.color || '#4f46e5' })
  showModal.value = true
}

async function save() {
  if (!form.name.trim()) {
    message.warning('Name is required')
    return
  }
  try {
    if (editing.value) {
      await categoryApi.update(editing.value.id, { ...form })
      message.success('Category updated')
    } else {
      await categoryApi.create({ ...form })
      message.success('Category created')
    }
    showModal.value = false
    await load()
    await meta.refreshCategories()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Save failed')
  }
}

async function remove(row: Category) {
  try {
    await categoryApi.remove(row.id)
    message.success('Category deleted')
    await load()
    await meta.refreshCategories()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Delete failed')
  }
}

const columns: DataTableColumns<Category> = [
  {
    title: 'Name',
    key: 'name',
    render: (row) =>
      h('div', { class: 'cat' }, [
        h('span', { class: 'cat__dot', style: { background: row.color || '#c7c9cf' } }),
        h('span', null, row.name),
      ]),
  },
  { title: 'Slug', key: 'slug', render: (row) => h('code', { class: 'pkb-code' }, row.slug) },
  { title: 'Description', key: 'description', ellipsis: { tooltip: true } },
  {
    title: 'Problems',
    key: 'problem_count',
    width: 100,
    render: (row) => h(NTag, { size: 'small', bordered: false }, { default: () => row.problem_count }),
  },
  {
    title: '',
    key: 'actions',
    width: 130,
    render: (row) =>
      h('div', { class: 'actions' }, [
        h(NButton, { size: 'tiny', quaternary: true, onClick: () => openEdit(row) }, { default: () => 'Edit' }),
        h(
          NPopconfirm,
          { onPositiveClick: () => remove(row) },
          {
            trigger: () =>
              h(NButton, { size: 'tiny', quaternary: true, type: 'error' }, { default: () => 'Delete' }),
            default: () => 'Delete this category?',
          },
        ),
      ]),
  },
]

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="Categories" subtitle="จัดกลุ่มปัญหาตามพื้นที่/ระบบ">
      <template #actions>
        <NButton size="small" type="primary" @click="openCreate">+ New Category</NButton>
      </template>
    </PageHeader>

    <NDataTable :loading="loading" :columns="columns" :data="rows" :bordered="false" />

    <NModal
      v-model:show="showModal"
      preset="card"
      :title="editing ? 'Edit category' : 'New category'"
      style="max-width: 440px"
    >
      <div class="form">
        <label>
          <span>Name</span>
          <NInput v-model:value="form.name" placeholder="Network" />
        </label>
        <label>
          <span>Description</span>
          <NInput
            v-model:value="form.description"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 4 }"
          />
        </label>
        <label>
          <span>Colour</span>
          <NColorPicker v-model:value="form.color" :show-alpha="false" :modes="['hex']" />
        </label>
        <NButton type="primary" block @click="save">{{ editing ? 'Save' : 'Create' }}</NButton>
      </div>
    </NModal>
  </div>
</template>

<style scoped>
.form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.form label {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: #6b7280;
}
:deep(.cat) {
  display: flex;
  align-items: center;
  gap: 8px;
}
:deep(.cat__dot) {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}
:deep(.actions) {
  display: flex;
  gap: 4px;
}
</style>
