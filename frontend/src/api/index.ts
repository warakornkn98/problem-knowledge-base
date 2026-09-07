import { request } from './client'
import type {
  AuthResult,
  Category,
  CommonError,
  DashboardBucket,
  DashboardSummary,
  EnumMeta,
  Paginated,
  Problem,
  ProblemListItem,
  ProblemQuery,
  ProblemWritePayload,
  RelatedLink,
  RelationType,
  SearchResult,
  Step,
  Tag,
  User,
} from './types'

export const authApi = {
  login: (login: string, password: string) =>
    request<AuthResult>({ method: 'POST', url: '/auth/login', data: { login, password } }),
  register: (data: { username: string; email: string; password: string; display_name?: string }) =>
    request<AuthResult>({ method: 'POST', url: '/auth/register', data }),
  me: () => request<User>({ method: 'GET', url: '/auth/me' }),
}

export const metaApi = {
  enums: () => request<EnumMeta>({ method: 'GET', url: '/meta/enums' }),
}

export const categoryApi = {
  list: () => request<Category[]>({ method: 'GET', url: '/categories' }),
  create: (data: { name: string; description?: string; color?: string }) =>
    request<Category>({ method: 'POST', url: '/categories', data }),
  update: (id: number, data: { name: string; description?: string; color?: string }) =>
    request<Category>({ method: 'PUT', url: `/categories/${id}`, data }),
  remove: (id: number) => request<null>({ method: 'DELETE', url: `/categories/${id}` }),
}

export const tagApi = {
  list: (q?: string) => request<Tag[]>({ method: 'GET', url: '/tags', params: { q } }),
  create: (name: string) => request<Tag>({ method: 'POST', url: '/tags', data: { name } }),
  update: (id: number, name: string) => request<Tag>({ method: 'PUT', url: `/tags/${id}`, data: { name } }),
  remove: (id: number) => request<null>({ method: 'DELETE', url: `/tags/${id}` }),
}

export const problemApi = {
  list: (params: ProblemQuery) =>
    request<Paginated<ProblemListItem>>({ method: 'GET', url: '/problems', params }),
  projects: () => request<string[]>({ method: 'GET', url: '/problems/projects' }),
  get: (id: number) => request<Problem>({ method: 'GET', url: `/problems/${id}` }),
  create: (data: ProblemWritePayload) => request<Problem>({ method: 'POST', url: '/problems', data }),
  update: (id: number, data: ProblemWritePayload) =>
    request<Problem>({ method: 'PUT', url: `/problems/${id}`, data }),
  remove: (id: number) => request<null>({ method: 'DELETE', url: `/problems/${id}` }),
  setStatus: (id: number, status: string) =>
    request<Problem>({ method: 'PATCH', url: `/problems/${id}/status`, data: { status } }),
  similar: (id: number, limit = 6) =>
    request<ProblemListItem[]>({ method: 'GET', url: `/problems/${id}/similar`, params: { limit } }),

  listSteps: (id: number) => request<Step[]>({ method: 'GET', url: `/problems/${id}/steps` }),
  addStep: (id: number, data: { action: string; result: string }) =>
    request<Step>({ method: 'POST', url: `/problems/${id}/steps`, data }),
  updateStep: (id: number, stepId: number, data: { action: string; result: string; step_no?: number }) =>
    request<Step>({ method: 'PUT', url: `/problems/${id}/steps/${stepId}`, data }),
  deleteStep: (id: number, stepId: number) =>
    request<null>({ method: 'DELETE', url: `/problems/${id}/steps/${stepId}` }),
  reorderSteps: (id: number, orderedIds: number[]) =>
    request<Step[]>({ method: 'PUT', url: `/problems/${id}/steps/reorder`, data: { ordered_ids: orderedIds } }),

  listRelated: (id: number) => request<RelatedLink[]>({ method: 'GET', url: `/problems/${id}/related` }),
  addRelated: (id: number, relatedProblemId: number, relationType: RelationType) =>
    request<RelatedLink>({
      method: 'POST',
      url: `/problems/${id}/related`,
      data: { related_problem_id: relatedProblemId, relation_type: relationType },
    }),
  removeRelated: (id: number, relatedId: number) =>
    request<null>({ method: 'DELETE', url: `/problems/${id}/related/${relatedId}` }),
}

export const searchApi = {
  search: (params: ProblemQuery) => request<SearchResult>({ method: 'GET', url: '/search', params }),
}

export const dashboardApi = {
  summary: () => request<DashboardSummary>({ method: 'GET', url: '/dashboard/summary' }),
  categories: () => request<DashboardBucket[]>({ method: 'GET', url: '/dashboard/categories' }),
  projects: () => request<DashboardBucket[]>({ method: 'GET', url: '/dashboard/projects' }),
  commonErrors: () => request<CommonError[]>({ method: 'GET', url: '/dashboard/common-errors' }),
}
