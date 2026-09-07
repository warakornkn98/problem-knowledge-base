<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NCollapse,
  NCollapseItem,
  NForm,
  NFormItem,
  NGrid,
  NGi,
  NInput,
  NSelect,
  NSpin,
  useMessage,
} from 'naive-ui'
import { problemApi } from '@/api'
import type { ProblemWritePayload } from '@/api/types'
import { ApiError } from '@/api/client'
import { useMetaStore } from '@/stores/meta'
import PageHeader from '@/components/PageHeader.vue'
import TagSelect from '@/components/TagSelect.vue'
import StepsEditor, { type DraftStep } from './components/StepsEditor.vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const meta = useMetaStore()

const id = computed(() => (route.params.id ? Number(route.params.id) : null))
const isEdit = computed(() => id.value !== null)
const loading = ref(false)
const saving = ref(false)

const form = reactive({
  title: '',
  description: '',
  error_message: '',
  project: '',
  environment: 'PROD',
  category_id: null as number | null,
  severity: 'MEDIUM',
  status: 'OPEN',
  root_cause: '',
  solution: '',
  prevention: '',
  tags: [] as string[],
})
const steps = ref<DraftStep[]>([])

const categoryOptions = computed(() => meta.categories.map((c) => ({ label: c.name, value: c.id })))
const projectOptions = computed(() => meta.projects.map((p) => ({ label: p, value: p })))
const enumOptions = (values: string[]) => values.map((v) => ({ label: v, value: v }))

onMounted(async () => {
  await meta.load().catch(() => {})
  if (isEdit.value && id.value !== null) {
    loading.value = true
    try {
      const p = await problemApi.get(id.value)
      Object.assign(form, {
        title: p.title,
        description: p.description,
        error_message: p.error_message,
        project: p.project,
        environment: p.environment,
        category_id: p.category_id,
        severity: p.severity,
        status: p.status,
        root_cause: p.root_cause,
        solution: p.solution,
        prevention: p.prevention,
        tags: p.tags.map((t) => t.name),
      })
      steps.value = p.steps.map((s) => ({ action: s.action, result: s.result }))
    } catch {
      message.error('Could not load the problem')
      router.push({ name: 'problems' })
    } finally {
      loading.value = false
    }
  }
})

function validate(): string | null {
  if (!form.title.trim()) return 'Title is required'
  if (!form.project.trim()) return 'Project is required'
  if (!form.environment) return 'Environment is required'
  if (!form.category_id) return 'Category is required'
  return null
}

async function save() {
  const problem = validate()
  if (problem) {
    message.warning(problem)
    return
  }
  saving.value = true
  const payload: ProblemWritePayload = {
    title: form.title.trim(),
    description: form.description,
    error_message: form.error_message,
    category_id: form.category_id as number,
    severity: form.severity as ProblemWritePayload['severity'],
    status: form.status as ProblemWritePayload['status'],
    environment: form.environment as ProblemWritePayload['environment'],
    project: form.project.trim(),
    root_cause: form.root_cause,
    solution: form.solution,
    prevention: form.prevention,
    tags: form.tags,
    steps: steps.value.filter((s) => s.action.trim()),
  }
  try {
    const saved = isEdit.value
      ? await problemApi.update(id.value as number, payload)
      : await problemApi.create(payload)
    await meta.load(true).catch(() => {})
    message.success(isEdit.value ? 'Problem updated' : 'Problem created')
    router.push({ name: 'problem-detail', params: { id: saved.id } })
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Save failed')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <PageHeader :title="isEdit ? 'Edit Problem' : 'New Problem'">
      <template #actions>
        <NButton size="small" @click="router.back()">Cancel</NButton>
        <NButton size="small" type="primary" :loading="saving" @click="save">
          {{ isEdit ? 'Save changes' : 'Create problem' }}
        </NButton>
      </template>
    </PageHeader>

    <NSpin :show="loading">
      <NForm label-placement="top">
        <NCard title="Basic information" size="small" class="section">
          <NFormItem label="Title" required>
            <NInput v-model:value="form.title" placeholder="e.g. Wazuh Server ส่ง Webhook แล้ว Timeout" />
          </NFormItem>
          <NGrid :cols="2" :x-gap="12" responsive="screen" item-responsive>
            <NGi span="2 m:1">
              <NFormItem label="Project" required>
                <NSelect
                  v-model:value="form.project"
                  filterable
                  tag
                  :options="projectOptions"
                  placeholder="Wazuh Backend"
                />
              </NFormItem>
            </NGi>
            <NGi span="2 m:1">
              <NFormItem label="Category" required>
                <NSelect v-model:value="form.category_id" :options="categoryOptions" placeholder="Network" />
              </NFormItem>
            </NGi>
            <NGi span="2 m:1">
              <NFormItem label="Environment" required>
                <NSelect v-model:value="form.environment" :options="enumOptions(meta.enums.environments)" />
              </NFormItem>
            </NGi>
            <NGi span="2 m:1">
              <NFormItem label="Severity">
                <NSelect v-model:value="form.severity" :options="enumOptions(meta.enums.severities)" />
              </NFormItem>
            </NGi>
            <NGi span="2 m:1">
              <NFormItem label="Status">
                <NSelect v-model:value="form.status" :options="enumOptions(meta.enums.statuses)" />
              </NFormItem>
            </NGi>
          </NGrid>
          <NFormItem label="Description">
            <NInput
              v-model:value="form.description"
              type="textarea"
              :autosize="{ minRows: 2, maxRows: 6 }"
              placeholder="สรุปสั้น ๆ ว่าปัญหาคืออะไร เกิดกับอะไร"
            />
          </NFormItem>
          <NFormItem label="Error message">
            <NInput
              v-model:value="form.error_message"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 12 }"
              placeholder="วาง error / stack trace / log ที่นี่ — จะแสดงเป็น code block"
              class="mono"
            />
          </NFormItem>
          <NFormItem label="Tags">
            <TagSelect v-model="form.tags" />
          </NFormItem>
        </NCard>

        <NCard title="Analysis" size="small" class="section">
          <p class="pkb-muted hint">
            เว้นว่างไว้ก่อนได้ ถ้าเป็น Quick Add แล้วค่อยกลับมาเติมทีหลัง
          </p>
          <NFormItem label="Root cause">
            <NInput v-model:value="form.root_cause" type="textarea" :autosize="{ minRows: 2, maxRows: 8 }" />
          </NFormItem>
          <NFormItem label="Solution">
            <NInput v-model:value="form.solution" type="textarea" :autosize="{ minRows: 2, maxRows: 8 }" />
          </NFormItem>
          <NFormItem label="Prevention">
            <NInput v-model:value="form.prevention" type="textarea" :autosize="{ minRows: 2, maxRows: 6 }" />
          </NFormItem>
        </NCard>

        <NCard size="small" class="section">
          <NCollapse :default-expanded-names="steps.length ? ['steps'] : []">
            <NCollapseItem :title="`Troubleshooting steps (${steps.length})`" name="steps">
              <StepsEditor v-model="steps" />
            </NCollapseItem>
          </NCollapse>
        </NCard>
      </NForm>
    </NSpin>
  </div>
</template>

<style scoped>
.section {
  margin-bottom: 16px;
}
.hint {
  font-size: 12px;
  margin: 0 0 12px;
}
:deep(.mono textarea) {
  font-family: var(--pkb-mono);
  font-size: 13px;
}
</style>
