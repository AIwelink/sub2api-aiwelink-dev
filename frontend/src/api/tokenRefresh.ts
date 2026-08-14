import axios from 'axios'
import type { ApiResponse } from '@/types'
import { getAPIBaseURL } from './url'
import {
  getInMemoryRefreshToken,
  setInMemoryRefreshToken
} from './authSecrets'

const AUTH_TOKEN_KEY = 'auth_token'
const AUTH_USER_KEY = 'auth_user'
const TOKEN_EXPIRES_AT_KEY = 'token_expires_at'
const TOKEN_REFRESH_TIMEOUT_MS = 30_000

export interface RefreshTokenResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface RefreshAuthTokensOptions {
  /** Access token attached to the request that received a 401 response. */
  failedAccessToken?: string | null
}

interface AuthSnapshot {
  accessToken: string | null
  refreshToken: string
  userID: number | null
}

let inFlightRefresh: Promise<RefreshTokenResponse> | null = null

function getStoredUserID(): number | null {
  const rawUser = localStorage.getItem(AUTH_USER_KEY)
  if (!rawUser) {
    return null
  }

  try {
    const id = Number((JSON.parse(rawUser) as { id?: unknown }).id)
    return Number.isFinite(id) && id > 0 ? id : null
  } catch {
    return null
  }
}

function readAuthSnapshot(): AuthSnapshot {
  const refreshToken = getInMemoryRefreshToken()
  if (!refreshToken) {
    throw new Error('No refresh token available')
  }

  return {
    accessToken: localStorage.getItem(AUTH_TOKEN_KEY),
    refreshToken,
    userID: getStoredUserID()
  }
}

function readCurrentTokenPair(failedAccessToken?: string | null): RefreshTokenResponse | null {
  const accessToken = localStorage.getItem(AUTH_TOKEN_KEY)
  const refreshToken = getInMemoryRefreshToken()
  const expiresAt = Number(localStorage.getItem(TOKEN_EXPIRES_AT_KEY))

  if (
    !failedAccessToken ||
    accessToken === failedAccessToken ||
    !accessToken ||
    !refreshToken ||
    !Number.isFinite(expiresAt) ||
    expiresAt <= Date.now()
  ) {
    return null
  }

  return {
    access_token: accessToken,
    refresh_token: refreshToken,
    expires_in: Math.max(1, Math.ceil((expiresAt - Date.now()) / 1000)),
    token_type: 'Bearer'
  }
}

function persistTokenPair(tokens: RefreshTokenResponse): void {
  localStorage.setItem(AUTH_TOKEN_KEY, tokens.access_token)
  localStorage.setItem(TOKEN_EXPIRES_AT_KEY, String(Date.now() + tokens.expires_in * 1000))
  setInMemoryRefreshToken(tokens.refresh_token)
}

async function requestTokenPair(snapshot: AuthSnapshot): Promise<RefreshTokenResponse> {
  const response = await axios.post<ApiResponse<RefreshTokenResponse>>(
    `${getAPIBaseURL()}/auth/refresh`,
    { refresh_token: snapshot.refreshToken },
    { headers: { 'Content-Type': 'application/json' }, timeout: TOKEN_REFRESH_TIMEOUT_MS }
  )
  const payload = response.data
  if (payload.code !== 0 || !payload.data) {
    throw new Error(payload.message || 'Token refresh failed')
  }

  if (
    getInMemoryRefreshToken() !== snapshot.refreshToken ||
    getStoredUserID() !== snapshot.userID
  ) {
    throw new Error('Session changed during token refresh')
  }

  persistTokenPair(payload.data)
  return payload.data
}

async function runRefresh(options: RefreshAuthTokensOptions): Promise<RefreshTokenResponse> {
  const currentPair = readCurrentTokenPair(options.failedAccessToken)
  if (currentPair) {
    return currentPair
  }

  const snapshot = readAuthSnapshot()
  return requestTokenPair(snapshot)
}

/**
 * Refresh and persist the browser session.
 *
 * Calls in the same document share one promise. Refresh credentials are intentionally scoped to
 * this document and are never synchronized through persistent browser storage.
 */
export function refreshAuthTokens(
  options: RefreshAuthTokensOptions = {}
): Promise<RefreshTokenResponse> {
  if (inFlightRefresh) {
    return inFlightRefresh
  }

  const pending = runRefresh(options)
  inFlightRefresh = pending
  const clearPending = (): void => {
    if (inFlightRefresh === pending) {
      inFlightRefresh = null
    }
  }
  void pending.then(clearPending, clearPending)
  return pending
}
