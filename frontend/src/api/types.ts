// Shared response types mirroring the Go API.

export interface Envelope<T> {
  success: boolean
  data: T
  message: string | null
}

export interface Pagination {
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface Paginated<T> {
  items: T[]
  pagination: Pagination
}

export type Severity = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'
export type Status = 'OPEN' | 'INVESTIGATING' | 'SOLVED' | 'KNOWN'
export type Environment = 'LOCAL' | 'DEV' | 'UAT' | 'PROD'
export type RelationType = 'RELATED' | 'SIMILAR' | 'CAUSED_BY' | 'DUPLICATE' | 'WORKAROUND'

export interface EnumMeta {
  severities: Severity[]
  statuses: Status[]
  environments: Environment[]
  relation_types: RelationType[]
}

export interface User {
  id: number
  username: string
  email: string
  display_name: string
  role: 'ADMIN' | 'MEMBER'
}

export interface AuthResult {
  token: string
  user: User
}

export interface Category {
  id: number
  name: string
  slug: string
  description: string
  color: string
  problem_count: number
  created_at: string
  updated_at: string
}

export interface Tag {
  id: number
  name: string
  slug: string
  usage_count: number
}

export interface CategoryRef {
  id: number
  name: string
  slug: string
  color: string
}

export interface TagRef {
  id: number
  name: string
  slug: string
}

export interface UserRef {
  id: number
  username: string
  display_name: string
}

export interface Step {
  id: number
  step_no: number
  action: string
  result: string
  created_at: string
  updated_at: string
}

export interface ProblemListItem {
  id: number
  title: string
  project: string
  environment: Environment
  severity: Severity
  status: Status
  category: CategoryRef | null
  tags: TagRef[]
  created_at: string
  updated_at: string
  solved_at: string | null
  rank?: number
  headline?: string
}

export interface Problem {
  id: number
  title: string
  description: string
  error_message: string
  category: CategoryRef | null
  category_id: number
  severity: Severity
  status: Status
  environment: Environment
  project: string
  root_cause: string
  solution: string
  prevention: string
  tags: TagRef[]
  steps: Step[]
  created_by: UserRef | null
  updated_by: UserRef | null
  created_at: string
  updated_at: string
  solved_at: string | null
}

export interface RelatedLink {
  id: number
  relation_type: RelationType
  created_at: string
  problem: ProblemListItem | null
}

export interface FacetCount {
  value: string
  label: string
  count: number
}

export interface Facets {
  categories: FacetCount[]
  statuses: FacetCount[]
  severities: FacetCount[]
  environments: FacetCount[]
  projects: FacetCount[]
  tags: FacetCount[]
}

export interface SearchResult {
  items: ProblemListItem[]
  facets: Facets
  pagination: Pagination
  query: string
}

export interface DashboardSummary {
  total: number
  open: number
  investigating: number
  solved: number
  known: number
  today: number
  this_week: number
  unsolved: number
}

export interface DashboardBucket {
  label: string
  slug?: string
  count: number
}

export interface CommonError {
  error_message: string
  count: number
}

export interface ProblemQuery {
  q?: string
  project?: string[]
  category_id?: number[]
  environment?: string[]
  severity?: string[]
  status?: string[]
  tag?: string[]
  created_from?: string
  created_to?: string
  solved_from?: string
  solved_to?: string
  sort?: string
  page?: number
  limit?: number
}

export interface ProblemWritePayload {
  title: string
  description?: string
  error_message?: string
  category_id: number
  severity?: Severity
  status?: Status
  environment: Environment
  project: string
  root_cause?: string
  solution?: string
  prevention?: string
  tags?: string[]
  tag_ids?: number[]
  steps?: { action: string; result: string }[]
}
