<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NIcon, useMessage } from 'naive-ui'

const props = withDefaults(
  defineProps<{ text: string; size?: 'tiny' | 'small' | 'medium'; label?: string }>(),
  { size: 'small', label: 'Copy' },
)

const message = useMessage()
const copied = ref(false)

async function copy() {
  try {
    await navigator.clipboard.writeText(props.text)
    copied.value = true
    setTimeout(() => (copied.value = false), 1500)
  } catch {
    message.error('Could not copy to clipboard')
  }
}
</script>

<template>
  <NButton :size="size" tertiary :type="copied ? 'success' : 'default'" @click="copy">
    <template #icon>
      <NIcon>
        <svg viewBox="0 0 24 24" width="1em" height="1em">
          <path
            v-if="!copied"
            fill="currentColor"
            d="M16 1H4a2 2 0 0 0-2 2v14h2V3h12V1m3 4H8a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h11a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2m0 16H8V7h11v14Z"
          />
          <path v-else fill="currentColor" d="M9 16.17L4.83 12l-1.42 1.41L9 19L21 7l-1.41-1.41L9 16.17Z" />
        </svg>
      </NIcon>
    </template>
    {{ copied ? 'Copied' : label }}
  </NButton>
</template>
