function envOr(key: string, fallback: string): string {
  const v = process.env[key]
  return v && v.trim() !== '' ? v.trim() : fallback
}

export class BrainClient {
  readonly baseUrl: string
  readonly username: string
  readonly password: string
  private token: string | null = null

  constructor(opts?: { baseUrl?: string; username?: string; password?: string }) {
    this.baseUrl = (opts?.baseUrl ?? envOr('AGENT_BRAIN_API_URL', 'http://127.0.0.1:8220')).replace(
      /\/$/,
      '',
    )
    this.username = opts?.username ?? envOr('AGENT_BRAIN_USERNAME', 'armin')
    this.password = opts?.password ?? envOr('AGENT_BRAIN_PASSWORD', 'dopadopa123')
  }

  async health() {
    const res = await fetch(`${this.baseUrl}/health`)
    if (!res.ok) throw new Error(`health ${res.status}`)
    return res.json()
  }

  async login(username?: string, password?: string) {
    const res = await fetch(`${this.baseUrl}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: username ?? this.username,
        password: password ?? this.password,
      }),
    })
    if (!res.ok) throw new Error(`login ${res.status}: ${await res.text()}`)
    const data = (await res.json()) as { token: string; user: unknown }
    this.token = data.token
    return data
  }

  private async authed(path: string, init: RequestInit = {}) {
    if (!this.token) await this.login()
    const headers = new Headers(init.headers)
    headers.set('Authorization', `Bearer ${this.token}`)
    if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
    const res = await fetch(`${this.baseUrl}${path}`, { ...init, headers })
    if (res.status === 401) {
      await this.login()
      headers.set('Authorization', `Bearer ${this.token}`)
      const retry = await fetch(`${this.baseUrl}${path}`, { ...init, headers })
      if (!retry.ok) throw new Error(`${path} ${retry.status}: ${await retry.text()}`)
      return retry.json()
    }
    if (!res.ok) throw new Error(`${path} ${res.status}: ${await res.text()}`)
    return res.json()
  }

  storeKnowledge(body: unknown) {
    return this.authed('/api/v1/knowledge', { method: 'POST', body: JSON.stringify(body) })
  }

  listKnowledge(kind?: string, q?: string) {
    const params = new URLSearchParams()
    if (kind) params.set('kind', kind)
    if (q) params.set('q', q)
    const qs = params.toString() ? `?${params}` : ''
    return this.authed(`/api/v1/knowledge${qs}`)
  }

  getKnowledge(id: string) {
    return this.authed(`/api/v1/knowledge/${id}`)
  }

  searchKnowledge(body: unknown) {
    return this.authed('/api/v1/knowledge/search', { method: 'POST', body: JSON.stringify(body) })
  }

  deleteKnowledge(id: string) {
    return this.authed(`/api/v1/knowledge/${id}`, { method: 'DELETE' })
  }

  ingestSession(body: unknown) {
    return this.authed('/api/v1/sessions/ingest', {
      method: 'POST',
      body: JSON.stringify(body),
    })
  }

  listSessions(q?: string) {
    const qs = q ? `?q=${encodeURIComponent(q)}` : ''
    return this.authed(`/api/v1/sessions${qs}`)
  }

  getSession(uuid: string) {
    return this.authed(`/api/v1/sessions/${uuid}`)
  }

  getSessionThinking(uuid: string) {
    return this.authed(`/api/v1/sessions/${uuid}/thinking`)
  }

  getSessionTurns(uuid: string) {
    return this.authed(`/api/v1/sessions/${uuid}/turns`)
  }
}
