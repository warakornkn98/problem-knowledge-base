<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  items: { label: string; count: number; slug?: string }[]
  color?: string
  clickable?: boolean
}>()

const emit = defineEmits<{ select: [item: { label: string; count: number; slug?: string }] }>()

const max = computed(() => Math.max(1, ...props.items.map((i) => i.count)))
</script>

<template>
  <div class="bl">
    <div
      v-for="item in items"
      :key="item.label"
      class="bl__row"
      :class="{ 'bl__row--click': clickable }"
      @click="clickable && emit('select', item)"
    >
      <div class="bl__head">
        <span class="bl__label" :title="item.label">{{ item.label }}</span>
        <span class="bl__count">{{ item.count }}</span>
      </div>
      <div class="bl__track">
        <div
          class="bl__fill"
          :style="{ width: `${(item.count / max) * 100}%`, background: color ?? '#4f46e5' }"
        />
      </div>
    </div>
    <p v-if="!items.length" class="pkb-muted bl__empty">No data yet</p>
  </div>
</template>

<style scoped>
.bl {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.bl__row--click {
  cursor: pointer;
}
.bl__head {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  margin-bottom: 4px;
}
.bl__label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 80%;
}
.bl__count {
  font-variant-numeric: tabular-nums;
  color: #6b7280;
}
.bl__track {
  height: 8px;
  background: #eef0f3;
  border-radius: 999px;
  overflow: hidden;
}
.bl__fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.3s ease;
}
.bl__empty {
  font-size: 13px;
}
</style>
