import { beforeEach, describe, expect, it, vi } from 'vitest'
import axios from 'axios'

vi.mock('axios', () => ({
  default: {
    post: vi.fn()
  }
}))

const mockedPost = vi.mocked(axios.post)

function seedStoredSession(): void {
  localStorage.setItem('auth_token', 'old-access')
  localStorage.setItem('token_expires_at', String(Date.now() - 1))
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

describe('refreshAuthTokens', () => {
  beforeEach(() => {
    localStorage.clear()
    mockedPost.mockReset()
    vi.resetModules()
  })

  it('shares one refresh request between concurrent callers in the same document', async () => {
    seedStoredSession()
    const { setInMemoryRefreshToken } = await import('@/api/authSecrets')
    setInMemoryRefreshToken('old-refresh')
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
    expect((await import('@/api/authSecrets')).getInMemoryRefreshToken()).toBe('new-refresh')
    expect(localStorage.getItem('refresh_token')).toBeNull()
  })

  it('does not restore a different account that replaces the session in flight', async () => {
    seedStoredSession()
    const { setInMemoryRefreshToken } = await import('@/api/authSecrets')
    setInMemoryRefreshToken('old-refresh')
    let resolveRequest!: (value: ReturnType<typeof refreshedResponse>) => void
    mockedPost.mockImplementationOnce(
      () => new Promise((resolve) => {
        resolveRequest = resolve
      })
    )
    const { refreshAuthTokens } = await import('@/api/tokenRefresh')

    const pending = refreshAuthTokens({ failedAccessToken: 'old-access' })
    localStorage.setItem('auth_user', JSON.stringify({ id: 8, email: 'other@example.com' }))
    localStorage.setItem('auth_token', 'other-access')
    setInMemoryRefreshToken('other-refresh')
    resolveRequest(refreshedResponse())

    await expect(pending).rejects.toThrow('Session changed during token refresh')
    expect(localStorage.getItem('auth_token')).toBe('other-access')
    expect((await import('@/api/authSecrets')).getInMemoryRefreshToken()).toBe('other-refresh')
  })

  it('does not treat profile metadata changes as a session switch', async () => {
    seedStoredSession()
    const { setInMemoryRefreshToken } = await import('@/api/authSecrets')
    setInMemoryRefreshToken('old-refresh')
    let resolveRequest!: (value: ReturnType<typeof refreshedResponse>) => void
    mockedPost.mockImplementationOnce(
      () => new Promise((resolve) => {
        resolveRequest = resolve
      })
    )
    const { refreshAuthTokens } = await import('@/api/tokenRefresh')

    const pending = refreshAuthTokens({ failedAccessToken: 'old-access' })
    localStorage.setItem('auth_user', JSON.stringify({ id: 7, email: 'updated@example.com' }))
    resolveRequest(refreshedResponse())

    await expect(pending).resolves.toMatchObject({ access_token: 'new-access' })
    expect(localStorage.getItem('auth_token')).toBe('new-access')
  })
})
