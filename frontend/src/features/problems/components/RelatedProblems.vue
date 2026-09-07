<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import {
  NButton,
  NInputNumber,
  NModal,
  NPopconfirm,
  NSelect,
  NSpin,
  useMessage,
} from 'naive-ui'
import { problemApi } from '@/api'
import type { RelatedLink, RelationType } from '@/api/types'
import { useMetaStore } from '@/stores/meta'
import MetaTag from '@/components/MetaTag.vue'
import { ApiError } from '@/api/client'

const props = defineProps<{ problemId: number }>()

const meta = useMetaStore()
const message = useMessage()

const loading = ref(false)
const links = ref<RelatedLink[]>([])
const showAdd = ref(false)
const draft = ref<{ related_problem_id: number | null; relation_type: RelationType }>({
  related_problem_id: null,
  relation_type: 'RELATED',
})

async function load() {
  loading.value = true
  try {
    links.value = await problemApi.listRelated(props.problemId)
  } catch {
    links.value = []
  } finally {
    loading.value = false
  }
}

async function add() {
  if (!draft.value.related_problem_id) {
    message.warning('Enter the related problem id')
    return
  }
  try {
    await problemApi.addRelated(props.problemId, draft.value.related_problem_id, draft.value.relation_type)
    showAdd.value = false
    draft.value = { related_problem_id: null, relation_type: 'RELATED' }
    await load()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Could not link the problem')
  }
}

async function remove(relatedId: number) {
  try {
    await problemApi.removeRelated(props.problemId, relatedId)
    await load()
  } catch {
    message.error('Could not remove the link')
  }
}

onMounted(load)
watch(() => props.problemId, load)
</script>

<template>
  <NSpin :show="loading">
    <ul v-if="links.length" class="rel">
      <li v-for="link in links" :key="link.id" class="rel__row">
        <MetaTag kind="relation" :value="link.relation_type" />
        <RouterLink
          v-if="link.problem"
          :to="{ name: 'problem-detail', params: { id: link.problem.id } }"
          class="rel__title"
        >
          {{ link.problem.title }}
        </RouterLink>
        <span v-else class="pkb-muted">#{{ link.id }}</span>
        <NPopconfirm @positive-click="link.problem && remove(link.problem.id)">
          <template #trigger>
            <NButton size="tiny" quaternary type="error">✕</NButton>
          </template>
          Remove this link?
        </NPopconfirm>
      </li>
    </ul>
    <p v-else class="pkb-muted empty">ยังไม่มีปัญหาที่เชื่อมโยง</p>

    <NButton size="tiny" dashed block style="margin-top: 8px" @click="showAdd = true">+ Link a problem</NButton>

    <NModal
      v-model:show="showAdd"
      preset="card"
      title="Link related problem"
      style="max-width: 420px"
    >
      <div class="form">
        <label>
          <span>Related problem id</span>
          <NInputNumber v-model:value="draft.related_problem_id" :min="1" placeholder="e.g. 42" />
        </label>
        <label>
          <span>Relation</span>
          <NSelect
            v-model:value="draft.relation_type"
            :options="meta.enums.relation_types.map((r) => ({ label: r.replace(/_/g, ' '), value: r }))"
          />
        </label>
        <NButton type="primary" block @click="add">Link</NButton>
      </div>
    </NModal>
  </NSpin>
</template>

<style scoped>
.rel {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.rel__row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.rel__title {
  flex: 1;
  font-size: 13px;
  color: #4f46e5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rel__title:hover {
  text-decoration: underline;
}
.empty {
  font-size: 13px;
}
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
</style>
