import type { RegisterRequest } from '@/types'

type PendingAuthTokenField = 'pending_auth_token' | 'pending_oauth_token'

interface PendingAdoptionDecision {
  adopt_display_name?: boolean
  adopt_avatar?: boolean
}

export interface PendingRegistrationData extends RegisterRequest {
  pending_auth_token?: string
  pending_auth_token_field?: PendingAuthTokenField
  pending_provider?: string
  pending_redirect?: string
  pending_adoption_decision?: PendingAdoptionDecision
}

const LEGACY_STORAGE_KEY = 'register_data'
let pendingRegistrationData: PendingRegistrationData | null = null

function clonePendingRegistrationData(
  data: PendingRegistrationData
): PendingRegistrationData {
  return {
    ...data,
    pending_adoption_decision: data.pending_adoption_decision
      ? { ...data.pending_adoption_decision }
      : undefined
  }
}

function removeLegacyRegistrationData(): void {
  sessionStorage.removeItem(LEGACY_STORAGE_KEY)
}

export function setPendingRegistrationData(data: PendingRegistrationData): void {
  removeLegacyRegistrationData()
  pendingRegistrationData = clonePendingRegistrationData(data)
}

export function getPendingRegistrationData(): PendingRegistrationData | null {
  if (pendingRegistrationData) {
    return clonePendingRegistrationData(pendingRegistrationData)
  }

  const legacyData = sessionStorage.getItem(LEGACY_STORAGE_KEY)
  removeLegacyRegistrationData()
  if (!legacyData) {
    return null
  }

  try {
    const parsed = JSON.parse(legacyData) as PendingRegistrationData
    pendingRegistrationData = clonePendingRegistrationData(parsed)
    return clonePendingRegistrationData(parsed)
  } catch {
    return null
  }
}

export function clearPendingRegistrationCaptchaProof(): void {
  if (!pendingRegistrationData) {
    return
  }

  delete pendingRegistrationData.turnstile_token
  delete pendingRegistrationData.tencent_captcha_ticket
  delete pendingRegistrationData.tencent_captcha_randstr
}

export function clearPendingRegistrationData(): void {
  removeLegacyRegistrationData()
  pendingRegistrationData = null
}
