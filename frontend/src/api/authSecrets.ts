let refreshToken: string | null = null

export function setInMemoryRefreshToken(token: string): void {
  refreshToken = token || null
}

export function getInMemoryRefreshToken(): string | null {
  return refreshToken
}

export function clearInMemoryRefreshToken(): void {
  refreshToken = null
}
