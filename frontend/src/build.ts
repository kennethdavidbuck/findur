export const frontendBuildSha = import.meta.env.VITE_BUILD_SHA || 'development'

export const isFullGitSha = (value: unknown): value is string =>
  typeof value === 'string' && /^[0-9a-f]{40}$/.test(value)

