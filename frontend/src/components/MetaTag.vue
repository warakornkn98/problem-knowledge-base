<script setup lang="ts">
import { computed } from 'vue'
import { NTag } from 'naive-ui'

type Kind = 'severity' | 'status' | 'environment' | 'relation'

const props = defineProps<{ kind: Kind; value: string; size?: 'small' | 'medium' | 'large' }>()

type NaiveType = 'default' | 'error' | 'warning' | 'success' | 'info' | 'primary'

const MAPS: Record<Kind, Record<string, NaiveType>> = {
  severity: { LOW: 'default', MEDIUM: 'info', HIGH: 'warning', CRITICAL: 'error' },
  status: { OPEN: 'error', INVESTIGATING: 'warning', SOLVED: 'success', KNOWN: 'info' },
  environment: { LOCAL: 'default', DEV: 'info', UAT: 'warning', PROD: 'error' },
  relation: {
    RELATED: 'default',
    SIMILAR: 'info',
    CAUSED_BY: 'warning',
    DUPLICATE: 'error',
    WORKAROUND: 'success',
  },
}

const type = computed<NaiveType>(() => MAPS[props.kind]?.[props.value] ?? 'default')
const label = computed(() => props.value.replace(/_/g, ' '))
</script>

<template>
  <NTag :type="type" :size="size ?? 'small'" :bordered="false" round>{{ label }}</NTag>
</template>
