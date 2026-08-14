import { beforeEach, describe, expect, it } from 'vitest'
import {
  clearInMemoryRefreshToken,
  getInMemoryRefreshToken,
  setInMemoryRefreshToken
} from '@/api/authSecrets'

describe('authSecrets', () => {
  beforeEach(() => {
    clearInMemoryRefreshToken()
    localStorage.clear()
    sessionStorage.clear()
  })

  it('keeps refresh tokens out of browser storage', () => {
    setInMemoryRefreshToken('refresh-secret')

    expect(getInMemoryRefreshToken()).toBe('refresh-secret')
    expect(localStorage.getItem('refresh_token')).toBeNull()
    expect(sessionStorage.getItem('refresh_token')).toBeNull()
  })

  it('clears the in-memory refresh token', () => {
    setInMemoryRefreshToken('refresh-secret')
    clearInMemoryRefreshToken()

    expect(getInMemoryRefreshToken()).toBeNull()
  })
})
