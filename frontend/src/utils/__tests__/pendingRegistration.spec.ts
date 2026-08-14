import { beforeEach, describe, expect, it } from 'vitest'
import {
  clearPendingRegistrationCaptchaProof,
  clearPendingRegistrationData,
  getPendingRegistrationData,
  setPendingRegistrationData
} from '@/utils/pendingRegistration'

describe('pendingRegistration', () => {
  beforeEach(() => {
    clearPendingRegistrationData()
    sessionStorage.clear()
    localStorage.clear()
  })

  it('keeps passwords and captcha proofs out of browser storage', () => {
    setPendingRegistrationData({
      email: 'user@example.com',
      password: 'secret-123',
      turnstile_token: 'captcha-proof'
    })

    expect(getPendingRegistrationData()).toEqual({
      email: 'user@example.com',
      password: 'secret-123',
      turnstile_token: 'captcha-proof',
      pending_adoption_decision: undefined
    })
    expect(sessionStorage.getItem('register_data')).toBeNull()
    expect(localStorage.getItem('register_data')).toBeNull()
  })

  it('removes captcha proofs after the first send-code attempt', () => {
    setPendingRegistrationData({
      email: 'user@example.com',
      password: 'secret-123',
      turnstile_token: 'turnstile-proof'
    })

    clearPendingRegistrationCaptchaProof()

    expect(getPendingRegistrationData()).toEqual({
      email: 'user@example.com',
      password: 'secret-123',
      pending_adoption_decision: undefined
    })
  })

  it('migrates legacy session data once and immediately removes the stored copy', () => {
    sessionStorage.setItem(
      'register_data',
      JSON.stringify({ email: 'legacy@example.com', password: 'legacy-secret' })
    )

    expect(getPendingRegistrationData()).toEqual({
      email: 'legacy@example.com',
      password: 'legacy-secret',
      pending_adoption_decision: undefined
    })
    expect(sessionStorage.getItem('register_data')).toBeNull()
  })
})
