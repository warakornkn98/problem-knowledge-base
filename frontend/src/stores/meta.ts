import { defineStore } from 'pinia'
import { ref } from 'vue'
import { categoryApi, metaApi, problemApi, tagApi } from '@/api'
import type { Category, EnumMeta, Tag } from '@/api/types'

const DEFAULT_ENUMS: EnumMeta = {
  severities: ['LOW', 'MEDIUM', 'HIGH', 'CRITICAL'],
  statuses: ['OPEN', 'INVESTIGATING', 'SOLVED', 'KNOWN'],
  environments: ['LOCAL', 'DEV', 'UAT', 'PROD'],
  relation_types: ['RELATED', 'SIMILAR', 'CAUSED_BY', 'DUPLICATE', 'WORKAROUND'],
}

/** Caches reference data (enums, categories, tags, projects) shared across views. */
export const useMetaStore = defineStore('meta', () => {
  const enums = ref<EnumMeta>(DEFAULT_ENUMS)
  const categories = ref<Category[]>([])
  const tags = ref<Tag[]>([])
  const projects = ref<string[]>([])
  const loaded = ref(false)

  async function load(force = false) {
    if (loaded.value && !force) return
    const [e, c, t, p] = await Promise.all([
      metaApi.enums(),
      categoryApi.list(),
      tagApi.list(),
      problemApi.projects().catch(() => [] as string[]),
    ])
    enums.value = e
    categories.value = c
    tags.value = t
    projects.value = p
    loaded.value = true
  }

  async function refreshCategories() {
    categories.value = await categoryApi.list()
  }

  async function refreshTags() {
    tags.value = await tagApi.list()
  }

  return { enums, categories, tags, projects, loaded, load, refreshCategories, refreshTags }
})
