import { beforeEach, describe, expect, it, vi } from 'vitest'
import axios from 'axios'

vi.mock('axios', () => ({
  default: {
    post: vi.fn()
  }
}))

const mockedPost = vi.mocked(axios.post)

function seedStoredSession(overrides: Partial<Record<string, string>> = {}): void {
  localStorage.setItem('auth_token', overrides.auth_token || 'old-access')
  localStorage.setItem('token_expires_at', overrides.token_expires_at || String(Date.now() - 1))
  localStorage.setItem('auth_user', JSON.stringify({ id: 7, email: 'admin@example.com' }))
}

function refreshedResponse() {
  return {
    data: {
      code: 0,
      message: 'ok',
      data: {
        access_token: 'new-access',
        refresh_token: 'new-refresh',
        expires_in: 3600,
        token_type: 'Bearer'
      }
    }
  }
}

async function seedRefreshToken(token = 'old-refresh'): Promise<void> {
  const { setInMemoryRefreshToken } = await import('@/api/authSecrets')
  setInMemoryRefreshToken(token)
}

describe('refreshAuthTokens', () => {
  beforeEach(() => {
    localStorage.clear()
    mockedPost.mockReset()
    vi.resetModules()
  })

  it('shares one refresh request between concurrent callers in the same document', async () => {
    seedStoredSession()
    await seedRefreshToken()
    let resolveRequest!: (value: ReturnType<typeof refreshedResponse>) => void
    mockedPost.mockImplementationOnce(
      () => new Promise((resolve) => {
        resolveRequest = resolve
      })
    )
    const { refreshAuthTokens } = await import('@/api/tokenRefresh')

    const first = refreshAuthTokens({ failedAccessToken: 'old-access' })
    const second = refreshAuthTokens({ failedAccessToken: 'old-access' })

    expect(mockedPost).toHaveBeenCalledTimes(1)
    resolveRequest(refreshedResponse())

    await expect(first).resolves.toMatchObject({ access_token: 'new-access' })
    await expect(second).resolves.toMatchObject({ refresh_token: 'new-refresh' })
    const { getInMemoryRefreshToken } = await import('@/api/authSecrets')
    expect(getInMemoryRefreshToken()).toBe('new-refresh')
    expect(localStorage.getItem('refresh_token')).toBeNull()
  })

  it('reuses a token pair already refreshed by another request in this document', async () => {
    seedStoredSession()
    await seedRefreshToken('new-refresh')
    localStorage.setItem('auth_token', 'new-access')
    localStorage.setItem('token_expires_at', String(Date.now() + 3600_000))
    const { refreshAuthTokens } = await import('@/api/tokenRefresh')

    await expect(
      refreshAuthTokens({ failedAccessToken: 'old-access' })
    ).resolves.toMatchObject({
      access_token: 'new-access',
      refresh_token: 'new-refresh'
    })
    expect(mockedPost).not.toHaveBeenCalled()
  })

  it('does not restore a session that changed while refresh was in flight', async () => {
    seedStoredSession()
    await seedRefreshToken()
    let resolveRequest!: (value: ReturnType<typeof refreshedResponse>) => void
    mockedPost.mockImplementationOnce(
      () => new Promise((resolve) => {
        resolveRequest = resolve
      })
    )
    const { refreshAuthTokens } = await import('@/api/tokenRefresh')
    const { setInMemoryRefreshToken } = await import('@/api/authSecrets')

    const pending = refreshAuthTokens({ failedAccessToken: 'old-access' })
    localStorage.setItem('auth_user', JSON.stringify({ id: 8, email: 'other@example.com' }))
    localStorage.setItem('auth_token', 'other-access')
    setInMemoryRefreshToken('other-refresh')
    resolveRequest(refreshedResponse())

    await expect(pending).rejects.toThrow('Session changed during token refresh')
    expect(localStorage.getItem('auth_token')).toBe('other-access')
    expect(localStorage.getItem('refresh_token')).toBeNull()
  })

  it('does not restore a session that logged out while refresh was in flight', async () => {
    seedStoredSession()
    await seedRefreshToken()
    let resolveRequest!: (value: ReturnType<typeof refreshedResponse>) => void
    mockedPost.mockImplementationOnce(
      () => new Promise((resolve) => {
        resolveRequest = resolve
      })
    )
    const { refreshAuthTokens } = await import('@/api/tokenRefresh')
    const { clearInMemoryRefreshToken } = await import('@/api/authSecrets')

    const pending = refreshAuthTokens({ failedAccessToken: 'old-access' })
    localStorage.clear()
    clearInMemoryRefreshToken()
    resolveRequest(refreshedResponse())

    await expect(pending).rejects.toThrow('Session changed during token refresh')
    expect(localStorage.getItem('auth_token')).toBeNull()
  })

  it('fails without a refresh credential in the current document', async () => {
    seedStoredSession()
    const { refreshAuthTokens } = await import('@/api/tokenRefresh')

    await expect(refreshAuthTokens()).rejects.toThrow('No refresh token available')
  })
})
