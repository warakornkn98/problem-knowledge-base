import axios, {
  AxiosError,
  type AxiosInstance,
  type AxiosRequestConfig,
} from 'axios'
import type { Envelope } from './types'

const TOKEN_KEY = 'pkb.token'

export function getToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export function setToken(token: string | null): void {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token)
    else localStorage.removeItem(TOKEN_KEY)
  } catch {
    /* ignore private-mode storage errors */
  }
}

/** Raised for any non-2xx API response; carries the server's message. */
export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

/** Callback invoked whenever the API returns 401, so the app can log out. */
let onUnauthorized: (() => void) | null = null
export function setUnauthorizedHandler(fn: () => void): void {
  onUnauthorized = fn
}

const http: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 20000,
  // Repeat array params as ?tag=a&tag=b (matches the Go handler).
  paramsSerializer: {
    serialize: (params) => {
      const usp = new URLSearchParams()
      for (const [key, value] of Object.entries(params)) {
        if (value === undefined || value === null || value === '') continue
        if (Array.isArray(value)) {
          for (const v of value) if (v !== undefined && v !== null && v !== '') usp.append(key, String(v))
        } else {
          usp.append(key, String(value))
        }
      }
      return usp.toString()
    },
  },
})

http.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (response) => response,
  (error: AxiosError<Envelope<unknown>>) => {
    const status = error.response?.status ?? 0
    const message =
      error.response?.data?.message ||
      error.message ||
      'Request failed. Please try again.'

    if (status === 401 && onUnauthorized) {
      onUnauthorized()
    }
    return Promise.reject(new ApiError(message, status))
  },
)

/** Unwraps the `{ success, data }` envelope and returns `data`. */
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const res = await http.request<Envelope<T>>(config)
  return res.data.data
}

export default http
