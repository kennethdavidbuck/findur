import { createContext, type ReactNode, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { useI18n, type Locale } from './i18n'
import { getDisplayPreferences, putDisplayPreferences, ProfileConflictError, ProfileDefenseError, ProfileSessionExpiredError } from './profile'
import { useTheme, type ThemePreference } from './theme'

type SaveState = 'loading' | 'saving' | 'saved' | 'retry'
type Value = { state: SaveState; setLocale: (value: Locale) => void; setTheme: (value: ThemePreference) => void }
const Context = createContext<Value | null>(null)

export function AuthenticatedPreferences({ children, onSessionExpired }: { children: ReactNode; onSessionExpired: () => void }) {
  const { locale, setLocale } = useI18n(); const { preference: theme, setPreference: setTheme } = useTheme()
  const [state, setState] = useState<SaveState>('loading'); const version = useRef(0); const latest = useRef({ locale, theme })
  const actions = useRef({ onSessionExpired, setLocale, setTheme })
  useEffect(() => { actions.current = { onSessionExpired, setLocale, setTheme } }, [onSessionExpired, setLocale, setTheme])
  useEffect(() => { latest.current = { locale, theme } }, [locale, theme])
  useEffect(() => { let active = true; void getDisplayPreferences().then(async (saved) => {
    if (!active) return
    if (saved === null) {
      try { const created = await putDisplayPreferences({ ...latest.current, expectedVersion: 0 }); if (active) { version.current = created.version; setState('saved') } } catch { if (active) setState('retry') }
      return
    }
    if ((saved.locale !== 'en' && saved.locale !== 'fr') || !['system', 'light', 'dark'].includes(saved.theme)) throw new Error('invalid display preferences response')
    version.current = saved.version; actions.current.setLocale(saved.locale); actions.current.setTheme(saved.theme); setState('saved')
  }).catch(async (error: unknown) => {
    if (!active) return
    if (error instanceof ProfileSessionExpiredError || error instanceof ProfileDefenseError) { actions.current.onSessionExpired(); return }
    setState('retry')
  }); return () => { active = false } }, [])
  const save = (next: Partial<{ locale: Locale; theme: ThemePreference }>) => {
    const values = { ...latest.current, ...next }; latest.current = values
    if (next.locale) setLocale(next.locale); if (next.theme) setTheme(next.theme); setState('saving')
    void putDisplayPreferences({ ...values, expectedVersion: version.current }).then((saved) => { version.current = saved.version; setState('saved') }).catch(async (error: unknown) => {
      if (error instanceof ProfileSessionExpiredError || error instanceof ProfileDefenseError) { onSessionExpired(); return }
      if (error instanceof ProfileConflictError) { try { const current = await getDisplayPreferences(); if (current !== null) { version.current = current.version; setLocale(current.locale); setTheme(current.theme) }; setState('retry'); return } catch { /* retry status below */ } }
      setState('retry')
    })
  }
  const value = useMemo<Value>(() => ({ state, setLocale: (value) => save({ locale: value }), setTheme: (value) => save({ theme: value }) }), [state])
  return <Context.Provider value={value}>{children}</Context.Provider>
}

export function useAuthenticatedPreferences() { return useContext(Context) }
