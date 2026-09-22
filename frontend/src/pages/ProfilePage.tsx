import { type FormEvent, type ReactNode, type RefObject, useEffect, useRef, useState } from 'react'
import { useI18n } from '../i18n'
import { getPersonalProfile, type PersonalProfile, type PersonalProfileInput, ProfileConflictError, ProfileDefenseError, ProfileSessionExpiredError, ProfileValidationError, putPersonalProfile } from '../profile'
import { useTheme } from '../theme'
import { useAuthenticatedPreferences } from '../authenticated-preferences'

type Draft = Omit<PersonalProfileInput, 'adultAttested'> & { adultAttested: boolean }
type Field = Exclude<keyof Draft, 'expectedVersion'>
const avatars = ['aurora', 'cedar', 'ember', 'harbour', 'meadow', 'solstice'] as const
const profileFields: Field[] = ['displayName', 'adultAttested', 'locationKey', 'relationshipIntent', 'biography', 'avatarKey', 'locale', 'theme']

const text = {
  en: {
    route: 'Profile · Personal details', title: 'Make the profile yours.', intro: 'Add the human context that sits alongside your portfolio. You decide what appears before and after a mutual match.',
    loading: 'Loading your profile…', unavailable: 'We couldn’t load your profile.', retry: 'Try again', summary: 'Please review these fields:', required: 'Required', private: 'Private', pre: 'Before a match', post: 'After a mutual match',
    identityTitle: 'Start with the essentials', identityIntro: 'Your name and city help make each introduction feel grounded. Your city stays private.', storyTitle: 'Say what you’re looking for', storyIntro: 'Keep it direct and personal. This is the context people see before deciding whether to connect.', portraitTitle: 'Choose your portrait', portraitIntro: 'Your portrait stays obscured until you and another person both express interest.', settingsTitle: 'Set your display', settingsIntro: 'These private choices only change how Findur appears to you.', safetyTitle: 'Adult-only space',
    displayName: 'Display name', displayHelp: 'The name people will see.', adult: 'I confirm that I am 18 or older', adultHelp: 'We record your confirmation time, not your birth date.', location: 'City', locationHelp: 'Used later for approximate distance. Precise coordinates are never shared.', intent: 'Relationship intent', intentHelp: 'Choose the description that feels closest right now.', biography: 'Short biography', biographyHelp: 'A few sentences about what makes you, you.', avatar: 'Portrait', avatarHelp: 'Shown only after a mutual match.', locale: 'Language', theme: 'Theme', preferencesHelp: 'Private display preference.', save: 'Save profile', saving: 'Saving…', saved: 'Profile saved.', failure: 'We couldn’t save your profile. Your entries are still here.', conflict: 'Your profile changed elsewhere. Your entries are still here.', reload: 'Reload latest and reapply my entries', reapplied: 'The latest saved version is loaded and your entries were reapplied. Review them, then save again.', choose: 'Choose your city',
    longTerm: 'Long-term', longTermHelp: 'I’m looking for a committed relationship.', openLongTerm: 'Open to long-term', openLongTermHelp: 'I’m open to where a meaningful connection leads.', figuring: 'Figuring it out', figuringHelp: 'I’m meeting people without forcing an outcome.', system: 'System', light: 'Light', dark: 'Dark', error: 'Review this field.', portraitNames: ['Aurora', 'Cedar', 'Ember', 'Harbour', 'Meadow', 'Solstice'],
  },
  fr: {
    route: 'Profil · Renseignements personnels', title: 'Créez un profil à votre image.', intro: 'Ajoutez le contexte humain qui accompagne votre portefeuille. Vous choisissez ce qui apparaît avant et après un match mutuel.',
    loading: 'Chargement de votre profil…', unavailable: 'Nous n’avons pas pu charger votre profil.', retry: 'Réessayer', summary: 'Vérifiez ces champs :', required: 'Obligatoire', private: 'Privé', pre: 'Avant un match', post: 'Après un match mutuel',
    identityTitle: 'Commencez par l’essentiel', identityIntro: 'Votre nom et votre ville rendent chaque présentation plus concrète. Votre ville reste privée.', storyTitle: 'Dites ce que vous recherchez', storyIntro: 'Soyez direct et personnel. Ce contexte est visible avant que les autres décident de créer un lien.', portraitTitle: 'Choisissez votre portrait', portraitIntro: 'Votre portrait reste masqué jusqu’à ce que l’intérêt soit mutuel.', settingsTitle: 'Réglez votre affichage', settingsIntro: 'Ces choix privés changent uniquement la façon dont Findur s’affiche pour vous.', safetyTitle: 'Espace réservé aux adultes',
    displayName: 'Nom affiché', displayHelp: 'Le nom que les autres verront.', adult: 'Je confirme avoir 18 ans ou plus', adultHelp: 'Nous enregistrons l’heure de votre confirmation, pas votre date de naissance.', location: 'Ville', locationHelp: 'Utilisée plus tard pour une distance approximative. Les coordonnées précises ne sont jamais partagées.', intent: 'Type de relation', intentHelp: 'Choisissez la description qui vous correspond le mieux maintenant.', biography: 'Courte biographie', biographyHelp: 'Quelques phrases sur ce qui vous rend unique.', avatar: 'Portrait', avatarHelp: 'Visible seulement après un match mutuel.', locale: 'Langue', theme: 'Thème', preferencesHelp: 'Préférence d’affichage privée.', save: 'Enregistrer le profil personnel', saving: 'Enregistrement…', saved: 'Profil enregistré.', failure: 'Impossible d’enregistrer le profil. Vos saisies sont toujours ici.', conflict: 'Votre profil a changé ailleurs. Vos saisies sont toujours ici.', reload: 'Recharger la version récente et réappliquer mes saisies', reapplied: 'La version enregistrée la plus récente est chargée et vos saisies ont été réappliquées. Vérifiez-les, puis enregistrez de nouveau.', choose: 'Choisissez votre ville',
    longTerm: 'Relation à long terme', longTermHelp: 'Je recherche une relation engagée.', openLongTerm: 'Ouvert à une relation à long terme', openLongTermHelp: 'Je suis ouvert à ce qu’un lien significatif peut devenir.', figuring: 'Je ne sais pas encore', figuringHelp: 'Je rencontre des gens sans imposer de résultat.', system: 'Système', light: 'Clair', dark: 'Sombre', error: 'Vérifiez ce champ.', portraitNames: ['Aurore', 'Cèdre', 'Braise', 'Havre', 'Prairie', 'Solstice'],
  },
} as const

function emptyDraft(locale: 'en' | 'fr', theme: 'system' | 'light' | 'dark'): Draft {
  return { displayName: '', adultAttested: false, locationKey: '', relationshipIntent: 'long-term', biography: '', avatarKey: 'aurora', locale, theme, expectedVersion: 0 }
}

function profileDraft(profile: PersonalProfile): Draft {
  return { displayName: profile.displayName, adultAttested: true, locationKey: profile.locationKey, relationshipIntent: profile.relationshipIntent, biography: profile.biography, avatarKey: profile.avatarKey, locale: profile.locale, theme: profile.theme, expectedVersion: profile.version }
}

export function ProfilePage({ headingRef, onSessionExpired }: { headingRef: RefObject<HTMLHeadingElement | null>, onSessionExpired: () => void }) {
  const { locale, setLocale } = useI18n()
  const { preference, setPreference } = useTheme()
  const authenticatedPreferences = useAuthenticatedPreferences()
  const copy = text[locale]
  const [draft, setDraft] = useState<Draft>(() => emptyDraft(locale, preference))
  const [locations, setLocations] = useState<Awaited<ReturnType<typeof getPersonalProfile>>['locations']>([])
  const [loading, setLoading] = useState(true)
  const [loadFailed, setLoadFailed] = useState(false)
  const [saving, setSaving] = useState(false)
  const [status, setStatus] = useState<'idle' | 'saved' | 'invalid' | 'failed' | 'conflict' | 'reapplied'>('idle')
  const [errors, setErrors] = useState<Field[]>([])
  const summaryRef = useRef<HTMLDivElement>(null)
  const recoveryRef = useRef<HTMLButtonElement>(null)
  const saveRef = useRef<HTMLButtonElement>(null)
  const failureRef = useRef<HTMLDivElement>(null)
  const focusedHeading = useRef(false)
  const initial = useRef({ locale, preference, onSessionExpired, setLocale, setPreference })

  const applySnapshot = (snapshot: Awaited<ReturnType<typeof getPersonalProfile>>) => {
    setLocations(snapshot.locations)
    if (snapshot.profile) {
      setDraft(profileDraft(snapshot.profile))
    } else {
      setDraft(emptyDraft(initial.current.locale, initial.current.preference))
    }
  }

  const load = async () => {
    setLoading(true)
    setLoadFailed(false)
    try {
      const snapshot = await getPersonalProfile()
      applySnapshot(snapshot)
    } catch (error) {
      if (error instanceof ProfileSessionExpiredError || error instanceof ProfileDefenseError) onSessionExpired()
      else setLoadFailed(true)
    } finally { setLoading(false) }
  }

  useEffect(() => {
    let active = true
    void getPersonalProfile().then((snapshot) => {
      if (!active) return
      applySnapshot(snapshot)
    }).catch((error: unknown) => {
      if (!active) return
      if (error instanceof ProfileSessionExpiredError || error instanceof ProfileDefenseError) initial.current.onSessionExpired()
      else setLoadFailed(true)
    }).finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [])
  useEffect(() => {
    if (!loading && !focusedHeading.current) {
      focusedHeading.current = true
      headingRef.current?.focus()
    }
  }, [headingRef, loading])
  useEffect(() => {
    if (status === 'invalid' && errors.length > 0) summaryRef.current?.focus()
    else if (status === 'conflict') recoveryRef.current?.focus()
    else if (status === 'failed') failureRef.current?.focus()
    else if (status === 'reapplied') saveRef.current?.focus()
    else if (status === 'saved' && !saving) saveRef.current?.focus()
  }, [errors, saving, status])

  const update = <K extends Field>(field: K, value: Draft[K]) => {
    setDraft((current) => ({ ...current, [field]: value }))
    setStatus('idle')
    setErrors((current) => current.filter((item) => item !== field))
  }
  const choosePreference = (field: 'locale' | 'theme', value: Draft[typeof field]) => {
    if (field === 'locale') {
      if (authenticatedPreferences) authenticatedPreferences.setLocale(value as Draft['locale'])
      else setLocale(value as Draft['locale'])
    } else if (authenticatedPreferences) authenticatedPreferences.setTheme(value as Draft['theme'])
    else setPreference(value as Draft['theme'])
    setDraft((current) => ({ ...current, [field]: value }))
    setStatus('idle')
    setErrors((current) => current.filter((item) => item !== field))
  }
  const submit = async (event: FormEvent) => {
    event.preventDefault()
    const clientErrors = requiredErrors(draft)
    if (clientErrors.length > 0) {
      setErrors(clientErrors)
      setStatus('invalid')
      return
    }
    setSaving(true)
    setStatus('idle')
    setErrors([])
    try {
      const saved = await putPersonalProfile({ ...draft, adultAttested: true, locale, theme: preference })
      setDraft(profileDraft(saved))
      setStatus('saved')
    } catch (error) {
      if (error instanceof ProfileSessionExpiredError || error instanceof ProfileDefenseError) onSessionExpired()
      else if (error instanceof ProfileConflictError) setStatus('conflict')
      else if (error instanceof ProfileValidationError) {
        const fieldErrors = error.fields.filter(isField)
        if (fieldErrors.length > 0) {
          setErrors(fieldErrors)
          setStatus('invalid')
        } else setStatus('failed')
      } else setStatus('failed')
    } finally { setSaving(false) }
  }
  const reloadAndReapply = async () => {
    setSaving(true)
    try {
      const snapshot = await getPersonalProfile()
      setLocations(snapshot.locations)
      setDraft((current) => ({ ...current, expectedVersion: snapshot.profile?.version ?? 0 }))
      setStatus('reapplied')
    } catch (error) {
      if (error instanceof ProfileSessionExpiredError || error instanceof ProfileDefenseError) onSessionExpired()
      else setStatus('failed')
    } finally { setSaving(false) }
  }
  const invalid = (field: Field) => errors.includes(field)
  const fieldError = (field: Field) => invalid(field) ? <span id={`${field}-error`} className="profile-error"><span aria-hidden="true">◇</span> {copy.error}</span> : null

  if (loading) return <section className="profile-page profile-page--state"><h1 ref={headingRef} tabIndex={-1}>{copy.title}</h1><p role="status">{copy.loading}</p></section>
  if (loadFailed) return <section className="profile-page profile-page--state"><h1 ref={headingRef} tabIndex={-1}>{copy.title}</h1><p role="alert">{copy.unavailable}</p><button className="action action--primary" onClick={() => void load()}>{copy.retry}</button></section>

  return <section className="profile-page">
    <header className="profile-intro">
      <p className="profile-route"><span aria-hidden="true">03</span>{copy.route}</p>
      <h1 ref={headingRef} tabIndex={-1}>{copy.title}</h1>
      <p>{copy.intro}</p>
      <dl className="profile-visibility-key" aria-label={locale === 'fr' ? 'Guide de visibilité' : 'Visibility guide'}>
        <div><dt><span className="visibility-mark visibility-mark--pre" aria-hidden="true" />{copy.pre}</dt><dd>{locale === 'fr' ? 'Nom, intention et biographie' : 'Name, intent and biography'}</dd></div>
        <div><dt><span className="visibility-mark visibility-mark--post" aria-hidden="true" />{copy.post}</dt><dd>{locale === 'fr' ? 'Portrait sélectionné' : 'Chosen portrait'}</dd></div>
        <div><dt><span className="visibility-mark visibility-mark--private" aria-hidden="true" />{copy.private}</dt><dd>{locale === 'fr' ? 'Ville, âge et affichage' : 'City, age and display'}</dd></div>
      </dl>
    </header>

    {errors.length > 0 && <div ref={summaryRef} className="profile-error-summary" role="alert" tabIndex={-1}><strong>{copy.summary}</strong><ul>{errors.map((field) => <li key={field}><a href={`#${field}`}>{labelFor(field, copy)}</a></li>)}</ul></div>}

    <form className="profile-form" onSubmit={(event) => void submit(event)} noValidate>
      <fieldset className="profile-edit-fields" disabled={saving}>
      <ProfileSection number="01" title={copy.identityTitle} intro={copy.identityIntro}>
        <ProfileField id="displayName" label={copy.displayName} annotation={`${copy.required} · ${copy.pre}`} help={copy.displayHelp} error={fieldError('displayName')}><input id="displayName" value={draft.displayName} required aria-invalid={invalid('displayName')} aria-describedby={`displayName-help${invalid('displayName') ? ' displayName-error' : ''}`} onChange={(event) => update('displayName', event.target.value)} /></ProfileField>
        <ProfileField id="locationKey" label={copy.location} annotation={`${copy.required} · ${copy.private}`} help={copy.locationHelp} error={fieldError('locationKey')}><select id="locationKey" value={draft.locationKey} required aria-invalid={invalid('locationKey')} aria-describedby={`locationKey-help${invalid('locationKey') ? ' locationKey-error' : ''}`} onChange={(event) => update('locationKey', event.target.value)}><option value="">{copy.choose}</option>{locations.map((location) => <option key={location.key} value={location.key}>{locale === 'fr' ? `${location.cityFr}, ${location.provinceFr}` : `${location.cityEn}, ${location.provinceEn}`}</option>)}</select></ProfileField>
        <div className={`profile-attestation${invalid('adultAttested') ? ' profile-attestation--invalid' : ''}`}><label htmlFor="adultAttested"><input id="adultAttested" type="checkbox" aria-label={copy.adult} checked={draft.adultAttested} aria-invalid={invalid('adultAttested')} aria-describedby={`adultAttested-help${invalid('adultAttested') ? ' adultAttested-error' : ''}`} onChange={(event) => update('adultAttested', event.target.checked)} /><span><strong>{copy.adult}</strong><small id="adultAttested-help">{copy.adultHelp}</small></span><span className="profile-field__annotation">{copy.required} · {copy.private}</span></label>{fieldError('adultAttested')}</div>
      </ProfileSection>

      <ProfileSection number="02" title={copy.storyTitle} intro={copy.storyIntro}>
        <fieldset id="relationshipIntent" className="profile-field profile-choice-field" aria-invalid={invalid('relationshipIntent')}><legend><span>{copy.intent}</span><span className="profile-field__annotation">{copy.required} · {copy.pre}</span></legend><p id="relationshipIntent-help">{copy.intentHelp}</p><div className="profile-choice-list">{([
          ['long-term', copy.longTerm, copy.longTermHelp], ['open-to-long-term', copy.openLongTerm, copy.openLongTermHelp], ['figuring-it-out', copy.figuring, copy.figuringHelp],
        ] as const).map(([value, label, help]) => <label key={value} className="profile-choice"><input type="radio" name="relationshipIntent" value={value} checked={draft.relationshipIntent === value} aria-describedby={`relationshipIntent-help${invalid('relationshipIntent') ? ' relationshipIntent-error' : ''}`} onChange={() => update('relationshipIntent', value)} /><span><strong>{label}</strong><small>{help}</small></span><span className="choice-diamond" aria-hidden="true">◇</span></label>)}</div>{fieldError('relationshipIntent')}</fieldset>
        <ProfileField id="biography" label={copy.biography} annotation={`${copy.required} · ${copy.pre}`} help={copy.biographyHelp} error={fieldError('biography')}><textarea id="biography" value={draft.biography} required aria-invalid={invalid('biography')} aria-describedby={`biography-help biography-count${invalid('biography') ? ' biography-error' : ''}`} onChange={(event) => update('biography', event.target.value)} /><span id="biography-count" className="character-count">{codePointLength(draft.biography)}/500</span></ProfileField>
      </ProfileSection>

      <ProfileSection number="03" title={copy.portraitTitle} intro={copy.portraitIntro}>
        <fieldset id="avatarKey" className="profile-field profile-avatar-field" aria-invalid={invalid('avatarKey')}><legend><span>{copy.avatar}</span><span className="profile-field__annotation">{copy.required} · {copy.post}</span></legend><p id="avatarKey-help">{copy.avatarHelp}</p><div className="avatar-options">{avatars.map((avatar, index) => <label key={avatar}><input type="radio" name="avatar" value={avatar} checked={draft.avatarKey === avatar} aria-describedby={`avatarKey-help${invalid('avatarKey') ? ' avatarKey-error' : ''}`} onChange={() => update('avatarKey', avatar)} /><span className={`avatar avatar--${avatar}`} aria-hidden="true" /><span className="avatar-name">{copy.portraitNames[index]}</span><span className="avatar-selected" aria-hidden="true">◆</span></label>)}</div>{fieldError('avatarKey')}</fieldset>
      </ProfileSection>

      <ProfileSection number="04" title={copy.settingsTitle} intro={copy.settingsIntro}>
        <fieldset id="locale" className="profile-field profile-choice-field profile-preference-field" aria-invalid={invalid('locale')}>
          <legend><span>{copy.locale}</span><span className="profile-field__annotation">{copy.required} · {copy.private}</span></legend>
          <p id="locale-help">{copy.preferencesHelp} {authenticatedPreferences && 'Changes save automatically.'}</p>
          <div className="profile-preference-options">
            <label><input type="radio" name="profile-locale" value="en" checked={locale === 'en'} aria-describedby={`locale-help${invalid('locale') ? ' locale-error' : ''}`} onChange={() => choosePreference('locale', 'en')} /><span>English</span></label>
            <label><input type="radio" name="profile-locale" value="fr" checked={locale === 'fr'} aria-describedby={`locale-help${invalid('locale') ? ' locale-error' : ''}`} onChange={() => choosePreference('locale', 'fr')} /><span>Français</span></label>
          </div>
          {fieldError('locale')}
        </fieldset>
        <fieldset id="theme" className="profile-field profile-choice-field profile-preference-field" aria-invalid={invalid('theme')}>
          <legend><span>{copy.theme}</span><span className="profile-field__annotation">{copy.required} · {copy.private}</span></legend>
          <p id="theme-help">{copy.preferencesHelp} {authenticatedPreferences && 'Changes save automatically.'}</p>
          <div className="profile-preference-options profile-preference-options--theme">
            {(['system', 'light', 'dark'] as const).map((theme) => <label key={theme}><input type="radio" name="profile-theme" value={theme} checked={preference === theme} aria-describedby={`theme-help${invalid('theme') ? ' theme-error' : ''}`} onChange={() => choosePreference('theme', theme)} /><span>{copy[theme]}</span></label>)}
          </div>
          {fieldError('theme')}
        </fieldset>
      </ProfileSection>
      </fieldset>

      <footer className="profile-save-region"><div><span className="mono-label">{copy.safetyTitle}</span><p>{locale === 'fr' ? 'Vos choix de visibilité sont appliqués par Findur; votre ville et votre confirmation d’âge restent privées.' : 'Findur enforces these visibility choices; your city and age confirmation remain private.'}</p></div><button ref={saveRef} className="action action--primary" type="submit" disabled={saving}>{saving ? copy.saving : copy.save}</button></footer>
      <div ref={failureRef} className="profile-status" role={status === 'failed' || status === 'conflict' ? 'alert' : 'status'} tabIndex={status === 'failed' ? -1 : undefined}>{status === 'saved' ? <><span aria-hidden="true">◆</span> {copy.saved}</> : status === 'failed' ? copy.failure : status === 'conflict' ? <>{copy.conflict} <button ref={recoveryRef} type="button" className="text-link" disabled={saving} onClick={() => void reloadAndReapply()}>{copy.reload}</button></> : status === 'reapplied' ? copy.reapplied : null}</div>
    </form>
  </section>
}

function ProfileSection({ number, title, intro, children }: { number: string, title: string, intro: string, children: ReactNode }) {
  return <section className="profile-section"><header><span aria-hidden="true">{number}</span><div><h2>{title}</h2><p>{intro}</p></div></header><div className="profile-section__fields">{children}</div></section>
}

function ProfileField({ id, label, annotation, help, error, children }: { id: string, label: string, annotation: string, help: string, error?: ReactNode, children: ReactNode }) {
  return <div className="profile-field"><span className="profile-field__heading"><label htmlFor={id}>{label}</label><span className="profile-field__annotation">{annotation}</span></span>{children}<p id={`${id}-help`}>{help}</p>{error}</div>
}

function labelFor(field: Field, copy: typeof text.en | typeof text.fr) {
  return ({ displayName: copy.displayName, adultAttested: copy.adult, locationKey: copy.location, relationshipIntent: copy.intent, biography: copy.biography, avatarKey: copy.avatar, locale: copy.locale, theme: copy.theme })[field]
}

function isField(value: string): value is Field { return profileFields.includes(value as Field) }

function requiredErrors(draft: Draft): Field[] {
  const fields: Field[] = []
  const displayNameLength = codePointLength(draft.displayName.trim())
  if (displayNameLength < 1 || displayNameLength > 60) fields.push('displayName')
  if (!draft.adultAttested) fields.push('adultAttested')
  if (draft.locationKey.length === 0) fields.push('locationKey')
  const biographyLength = codePointLength(draft.biography.trim())
  if (biographyLength < 1 || biographyLength > 500) fields.push('biography')
  return fields
}

function codePointLength(value: string) { return Array.from(value).length }
