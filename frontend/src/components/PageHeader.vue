<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>()
</script>

<template>
  <div class="ph">
    <div class="ph__text">
      <h1 class="ph__title">{{ title }}</h1>
      <p v-if="subtitle" class="ph__subtitle pkb-muted">{{ subtitle }}</p>
    </div>
    <div class="ph__actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<style scoped>
.ph {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}
.ph__title {
  margin: 0;
  font-size: 22px;
  font-weight: 650;
}
.ph__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
}
.ph__actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
</style>
