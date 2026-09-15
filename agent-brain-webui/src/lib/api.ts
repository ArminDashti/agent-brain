const TOKEN_KEY = 'agent_brain_token'

export type KnowledgeKind = 'general_knowledge' | 'solution' | 'task_procedure' | 'expertise'
export type StoreTarget = 'postgres' | 'qdrant' | 'both'
export type SearchMode = 'structured' | 'semantic' | 'hybrid'

export type KnowledgeItem = {
  id: string
  kind: KnowledgeKind | string
  title: string
  body: string
  tags: string[]
  metadata: Record<string, unknown>
  in_postgres: boolean
  in_qdrant: boolean
  created_at: string
  updated_at: string
  score?: number
}

function apiBase(): string {
  return (import.meta.env.VITE_API_BASE as string | undefined)?.replace(/\/$/, '') || ''
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && init.body) headers.set('Content-Type', 'application/json')
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const res = await fetch(`${apiBase()}${path}`, { ...init, headers })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || res.statusText)
  }
  return res.json() as Promise<T>
}

export async function login(username: string, password: string) {
  const data = await request<{ token: string }>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
  setToken(data.token)
  return data
}

export function listKnowledge(kind?: string, q = '') {
  const params = new URLSearchParams()
  if (kind) params.set('kind', kind)
  if (q) params.set('q', q)
  const qs = params.toString() ? `?${params}` : ''
  return request<{ items: KnowledgeItem[] }>(`/api/v1/knowledge${qs}`)
}

export function createKnowledge(body: {
  kind: string
  title: string
  body: string
  tags?: string[]
  store_target: StoreTarget
}) {
  return request<KnowledgeItem>('/api/v1/knowledge', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function searchKnowledge(body: { query: string; kind?: string; mode: SearchMode }) {
  return request<{ items: KnowledgeItem[] }>('/api/v1/knowledge/search', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function deleteKnowledge(id: string) {
  return request<{ ok: boolean }>(`/api/v1/knowledge/${id}`, { method: 'DELETE' })
}

export type SessionSummary = {
  uuid: string
  name: string
  subtitle: string
  status: string
  unified_mode: string
  workspace_id: string
  workspace_path: string
  context_usage_percent: number | null
  input_tokens: number | null
  output_tokens: number | null
  cache_read_tokens: number | null
  cache_write_tokens: number | null
  model_config: unknown
  total_lines_added: number
  total_lines_removed: number
  files_changed_count: number
  turn_count?: number
  thinking_count?: number
}

export function listSessions(q = '') {
  const qs = q ? `?q=${encodeURIComponent(q)}` : ''
  return request<{ sessions: SessionSummary[] }>(`/api/v1/sessions${qs}`)
}

export function getSession(uuid: string) {
  return request<SessionSummary>(`/api/v1/sessions/${uuid}`)
}

export function getSessionThinking(uuid: string) {
  return request<{ thinking: Array<Record<string, unknown>> }>(`/api/v1/sessions/${uuid}/thinking`)
}

export function getSessionTurns(uuid: string) {
  return request<{ turns: Array<Record<string, unknown>> }>(`/api/v1/sessions/${uuid}/turns`)
}

export function fmtNum(v: number | null | undefined): string {
  if (v == null) return '—'
  return new Intl.NumberFormat().format(v)
}

export function fmtPct(v: number | null | undefined): string {
  if (v == null) return '—'
  return `${v.toFixed(1)}%`
}

export const KIND_LABELS: Record<string, string> = {
  general_knowledge: 'General Knowledge',
  solution: 'Solutions',
  task_procedure: 'Task Procedures',
  expertise: 'Expertise',
}
