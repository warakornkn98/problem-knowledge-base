<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NDropdown,
  NPopconfirm,
  NSpin,
  NTag,
  useMessage,
} from 'naive-ui'
import { problemApi } from '@/api'
import type { Problem } from '@/api/types'
import { ApiError } from '@/api/client'
import { useMetaStore } from '@/stores/meta'
import PageHeader from '@/components/PageHeader.vue'
import CodeBlock from '@/components/CodeBlock.vue'
import MetaTag from '@/components/MetaTag.vue'
import SimilarProblems from './components/SimilarProblems.vue'
import RelatedProblems from './components/RelatedProblems.vue'
import { formatDateTime, formatRelative } from '@/composables/format'

const props = defineProps<{ id: string }>()

const router = useRouter()
const message = useMessage()
const meta = useMetaStore()

const loading = ref(true)
const problem = ref<Problem | null>(null)
const problemId = computed(() => Number(props.id))

async function load() {
  loading.value = true
  try {
    problem.value = await problemApi.get(problemId.value)
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Problem not found')
    router.push({ name: 'problems' })
  } finally {
    loading.value = false
  }
}

const statusOptions = computed(() =>
  meta.enums.statuses.map((s) => ({ label: `Mark ${s}`, key: s })),
)

async function changeStatus(status: string) {
  if (!problem.value) return
  try {
    problem.value = await problemApi.setStatus(problem.value.id, status)
    message.success(`Status → ${status}`)
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Could not change status')
  }
}

async function remove() {
  if (!problem.value) return
  try {
    await problemApi.remove(problem.value.id)
    message.success('Problem deleted')
    router.push({ name: 'problems' })
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Delete failed')
  }
}

onMounted(load)
watch(problemId, load)
</script>

<template>
  <NSpin :show="loading">
    <div v-if="problem">
      <PageHeader :title="problem.title">
        <template #actions>
          <NDropdown :options="statusOptions" @select="changeStatus" trigger="click">
            <NButton size="small">Status: {{ problem.status }}</NButton>
          </NDropdown>
          <RouterLink :to="{ name: 'problem-edit', params: { id: problem.id } }">
            <NButton size="small" type="primary" ghost>Edit</NButton>
          </RouterLink>
          <NPopconfirm @positive-click="remove">
            <template #trigger>
              <NButton size="small" type="error" ghost>Delete</NButton>
            </template>
            Delete this problem permanently?
          </NPopconfirm>
        </template>
      </PageHeader>

      <div class="meta-row">
        <MetaTag kind="status" :value="problem.status" size="medium" />
        <MetaTag kind="severity" :value="problem.severity" size="medium" />
        <MetaTag kind="environment" :value="problem.environment" size="medium" />
        <NTag v-if="problem.category" size="small" :bordered="false">{{ problem.category.name }}</NTag>
        <span class="pkb-muted">·</span>
        <span class="pkb-muted">{{ problem.project }}</span>
      </div>

      <div class="layout">
        <div class="main">
          <NCard title="Description" size="small" class="block">
            <p v-if="problem.description" class="text">{{ problem.description }}</p>
            <p v-else class="pkb-muted">—</p>
          </NCard>

          <NCard title="Error message" size="small" class="block" content-style="padding: 12px">
            <CodeBlock :content="problem.error_message" title="error" empty-text="ไม่ได้บันทึก error message" />
          </NCard>

          <NCard title="Root cause" size="small" class="block">
            <p v-if="problem.root_cause" class="text">{{ problem.root_cause }}</p>
            <p v-else class="pkb-muted">ยังไม่ได้วิเคราะห์</p>
          </NCard>

          <NCard title="Solution" size="small" class="block">
            <p v-if="problem.solution" class="text">{{ problem.solution }}</p>
            <p v-else class="pkb-muted">ยังไม่มีวิธีแก้</p>
          </NCard>

          <NCard :title="`Troubleshooting (${problem.steps.length})`" size="small" class="block">
            <ol v-if="problem.steps.length" class="steps">
              <li v-for="step in problem.steps" :key="step.id" class="step">
                <div class="step__action">{{ step.action }}</div>
                <div v-if="step.result" class="step__result pkb-muted">→ {{ step.result }}</div>
              </li>
            </ol>
            <p v-else class="pkb-muted">ยังไม่มีขั้นตอน</p>
          </NCard>

          <NCard title="Prevention" size="small" class="block">
            <p v-if="problem.prevention" class="text">{{ problem.prevention }}</p>
            <p v-else class="pkb-muted">—</p>
          </NCard>
        </div>

        <aside class="side">
          <NCard title="Details" size="small" class="block">
            <dl class="facts">
              <dt>Tags</dt>
              <dd>
                <template v-if="problem.tags.length">
                  <RouterLink
                    v-for="t in problem.tags"
                    :key="t.id"
                    :to="{ name: 'problems', query: { tag: t.slug } }"
                  >
                    <NTag size="tiny" :bordered="false" style="margin: 0 4px 4px 0">{{ t.name }}</NTag>
                  </RouterLink>
                </template>
                <span v-else class="pkb-muted">—</span>
              </dd>
              <dt>Created</dt>
              <dd :title="formatDateTime(problem.created_at)">
                {{ formatRelative(problem.created_at) }}
                <span v-if="problem.created_by" class="pkb-muted">· {{ problem.created_by.display_name }}</span>
              </dd>
              <dt>Updated</dt>
              <dd :title="formatDateTime(problem.updated_at)">{{ formatRelative(problem.updated_at) }}</dd>
              <dt v-if="problem.solved_at">Solved</dt>
              <dd v-if="problem.solved_at">{{ formatDateTime(problem.solved_at) }}</dd>
            </dl>
          </NCard>

          <NCard title="Similar problems" size="small" class="block">
            <SimilarProblems :problem-id="problem.id" />
          </NCard>

          <NCard title="Related problems" size="small" class="block">
            <RelatedProblems :problem-id="problem.id" />
          </NCard>
        </aside>
      </div>
    </div>
  </NSpin>
</template>

<style scoped>
.meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}
.layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 16px;
}
@media (max-width: 900px) {
  .layout {
    grid-template-columns: 1fr;
  }
}
.block {
  margin-bottom: 14px;
}
.text {
  white-space: pre-wrap;
  line-height: 1.7;
  margin: 0;
}
.steps {
  margin: 0;
  padding-left: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.step__action {
  font-weight: 550;
}
.step__result {
  font-size: 13px;
  margin-top: 2px;
}
.facts {
  display: grid;
  grid-template-columns: 84px 1fr;
  gap: 8px 10px;
  margin: 0;
  font-size: 13px;
}
.facts dt {
  color: #8a8f98;
}
.facts dd {
  margin: 0;
}
</style>
