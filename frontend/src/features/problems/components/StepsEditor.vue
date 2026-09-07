<script setup lang="ts">
import { NButton, NInput } from 'naive-ui'

export interface DraftStep {
  action: string
  result: string
}

const model = defineModel<DraftStep[]>({ default: () => [] })

function add() {
  model.value = [...model.value, { action: '', result: '' }]
}
function remove(index: number) {
  model.value = model.value.filter((_, i) => i !== index)
}
function move(index: number, dir: -1 | 1) {
  const next = index + dir
  if (next < 0 || next >= model.value.length) return
  const copy = [...model.value]
  ;[copy[index], copy[next]] = [copy[next], copy[index]]
  model.value = copy
}
</script>

<template>
  <div class="steps">
    <div v-for="(step, i) in model" :key="i" class="step">
      <div class="step__no">{{ i + 1 }}</div>
      <div class="step__body">
        <NInput
          :value="step.action"
          placeholder="Action — e.g. ตรวจสอบ DNS / curl endpoint"
          @update:value="(v) => (model[i].action = v)"
        />
        <NInput
          :value="step.result"
          type="textarea"
          :autosize="{ minRows: 1, maxRows: 4 }"
          placeholder="Result — what happened"
          @update:value="(v) => (model[i].result = v)"
        />
      </div>
      <div class="step__actions">
        <NButton size="tiny" quaternary :disabled="i === 0" @click="move(i, -1)">↑</NButton>
        <NButton size="tiny" quaternary :disabled="i === model.length - 1" @click="move(i, 1)">↓</NButton>
        <NButton size="tiny" quaternary type="error" @click="remove(i)">✕</NButton>
      </div>
    </div>

    <NButton dashed block size="small" @click="add">+ Add step</NButton>
  </div>
</template>

<style scoped>
.steps {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.step {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}
.step__no {
  width: 24px;
  height: 24px;
  flex-shrink: 0;
  border-radius: 50%;
  background: #eef0f3;
  color: #4b5563;
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 6px;
}
.step__body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.step__actions {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
</style>
