<script setup lang="ts">
import { computed } from 'vue'
import { NSelect } from 'naive-ui'
import { useMetaStore } from '@/stores/meta'

// v-model is an array of tag *names* (lowercased). New names can be typed in.
const model = defineModel<string[]>({ default: () => [] })

defineProps<{ placeholder?: string }>()

const meta = useMetaStore()

const options = computed(() => {
  const known = meta.tags.map((t) => ({ label: t.name, value: t.name }))
  const extra = model.value
    .filter((name) => !meta.tags.some((t) => t.name === name))
    .map((name) => ({ label: name, value: name }))
  return [...extra, ...known]
})

function normalize(value: string[]) {
  model.value = Array.from(new Set(value.map((v) => v.trim().toLowerCase()).filter(Boolean)))
}
</script>

<template>
  <NSelect
    :value="model"
    multiple
    filterable
    tag
    clearable
    :options="options"
    :placeholder="placeholder ?? 'Add tags…'"
    @update:value="normalize"
  />
</template>
