<script setup lang="ts">
import { computed } from 'vue'
import { NButton, NCollapseTransition, NDatePicker, NInput, NSelect } from 'naive-ui'
import { useMetaStore } from '@/stores/meta'
import type { ProblemQuery } from '@/api/types'

const model = defineModel<ProblemQuery>({ required: true })
defineProps<{ expanded: boolean }>()
const emit = defineEmits<{ apply: []; reset: [] }>()

const meta = useMetaStore()

const categoryOptions = computed(() =>
  meta.categories.map((c) => ({ label: c.name, value: c.id })),
)
const tagOptions = computed(() => meta.tags.map((t) => ({ label: t.name, value: t.slug })))
const projectOptions = computed(() => meta.projects.map((p) => ({ label: p, value: p })))
const asOptions = (values: string[]) => values.map((v) => ({ label: v, value: v }))

const createdRange = computed<[number, number] | null>({
  get: () => rangeFrom(model.value.created_from, model.value.created_to),
  set: (v) => {
    model.value.created_from = v ? isoDate(v[0]) : undefined
    model.value.created_to = v ? isoDate(v[1]) : undefined
  },
})
const solvedRange = computed<[number, number] | null>({
  get: () => rangeFrom(model.value.solved_from, model.value.solved_to),
  set: (v) => {
    model.value.solved_from = v ? isoDate(v[0]) : undefined
    model.value.solved_to = v ? isoDate(v[1]) : undefined
  },
})

function rangeFrom(a?: string, b?: string): [number, number] | null {
  if (!a || !b) return null
  return [new Date(a).getTime(), new Date(b).getTime()]
}
function isoDate(ms: number) {
  return new Date(ms).toISOString().slice(0, 10)
}
</script>

<template>
  <div class="filters">
    <div class="filters__top">
      <NInput
        v-model:value="model.q"
        placeholder="Search error, title, root cause, solution, tags…"
        clearable
        @keyup.enter="emit('apply')"
      >
        <template #prefix>🔎</template>
      </NInput>
      <NButton type="primary" @click="emit('apply')">Search</NButton>
    </div>

    <NCollapseTransition :show="expanded">
      <div class="filters__grid">
        <label>
          <span>Project</span>
          <NSelect
            v-model:value="model.project"
            multiple
            filterable
            tag
            clearable
            :options="projectOptions"
            placeholder="Any project"
          />
        </label>
        <label>
          <span>Category</span>
          <NSelect v-model:value="model.category_id" multiple clearable :options="categoryOptions" placeholder="Any" />
        </label>
        <label>
          <span>Environment</span>
          <NSelect
            v-model:value="model.environment"
            multiple
            clearable
            :options="asOptions(meta.enums.environments)"
            placeholder="Any"
          />
        </label>
        <label>
          <span>Severity</span>
          <NSelect
            v-model:value="model.severity"
            multiple
            clearable
            :options="asOptions(meta.enums.severities)"
            placeholder="Any"
          />
        </label>
        <label>
          <span>Status</span>
          <NSelect
            v-model:value="model.status"
            multiple
            clearable
            :options="asOptions(meta.enums.statuses)"
            placeholder="Any"
          />
        </label>
        <label>
          <span>Tags</span>
          <NSelect v-model:value="model.tag" multiple filterable clearable :options="tagOptions" placeholder="Any" />
        </label>
        <label>
          <span>Created between</span>
          <NDatePicker v-model:value="createdRange" type="daterange" clearable />
        </label>
        <label>
          <span>Solved between</span>
          <NDatePicker v-model:value="solvedRange" type="daterange" clearable />
        </label>
      </div>
      <div class="filters__actions">
        <NButton size="small" @click="emit('reset')">Clear filters</NButton>
        <NButton size="small" type="primary" @click="emit('apply')">Apply</NButton>
      </div>
    </NCollapseTransition>
  </div>
</template>

<style scoped>
.filters {
  background: #fff;
  border: 1px solid #e3e5e8;
  border-radius: 10px;
  padding: 14px;
  margin-bottom: 16px;
}
.filters__top {
  display: flex;
  gap: 8px;
}
.filters__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
  margin-top: 14px;
}
.filters__grid label {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: #6b7280;
}
.filters__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 12px;
}
</style>
