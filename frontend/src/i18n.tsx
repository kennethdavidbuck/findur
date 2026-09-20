import { createContext, type ReactNode, useContext, useEffect, useMemo, useState } from 'react'

export type Locale = 'en' | 'fr'

const localeStorageKey = 'findur-locale'

const copy = {
  en: {
    localeName: 'English',
    skipLink: 'Skip to content',
    languageLabel: 'Language',
    themeLabel: 'Theme',
    themes: { system: 'System', light: 'Light', dark: 'Dark' },
    nav: {
      home: 'Home',
      about: 'About',
      primary: 'Public navigation',
    },
    owner: {
      action: 'Owner access',
      unavailable: 'Review the secure connection boundary.',
    },
    meta: {
      home: {
        title: 'Findur — Portfolio-first introductions',
        description:
          'Explore Findur, an 18+ evaluation concept for portfolio-first introductions and progressive identity reveal.',
      },
      about: {
        title: 'About Findur',
        description:
          'Learn why Findur explores portfolio context, deliberate disclosure, and progressive identity reveal.',
      },
      connect: {
        title: 'Connect with SnapTrade — Findur',
        description: 'Review the staged consent boundary before securely connecting SnapTrade to Findur.',
      },
    },
    home: {
      eyebrow: 'Portfolio-first introductions',
      titleBefore: 'Find a different pattern in the',
      titleAccent: 'same sky.',
      intro:
        'Findur explores compatibility through the shape of a portfolio before a photo enters the picture. Evidence opens the conversation; mutual interest reveals the person.',
      demoLabel: '18+ evaluation demo',
      aboutAction: 'Explore the idea',
      visualSummary:
        'Four connected ideas—foundation, curiosity, patience, and perspective—form a constellation. The illustration conveys relationships, not value or performance.',
      nodes: {
        foundation: 'Foundation',
        curiosity: 'Curiosity',
        patience: 'Patience',
        perspective: 'Perspective',
      },
      visualCaption: 'A portfolio is context, not a score.',
      visualNote: 'Illustrative pattern · no live financial data',
      howEyebrow: 'Concept flow · 01—03',
      howTitle: 'More context. Less performance.',
      steps: [
        {
          number: '01',
          title: 'Connect deliberately',
          body: 'A connected portfolio remains private by default. The concept starts with clear consent and selected accounts.',
        },
        {
          number: '02',
          title: 'Choose what is visible',
          body: 'Disclosure is a choice, not a status tier. A preview shows what another person could understand before anything is saved.',
        },
        {
          number: '03',
          title: 'Reveal identity together',
          body: 'A photo stays obscured until interest is mutual, keeping the first introduction focused on patterns and intent.',
        },
      ],
      boundaryEyebrow: 'Current boundary',
      boundaryTitle: 'A product demonstration, not a public dating launch.',
      boundaryBody:
        'The current build is designed to evaluate one responsible product direction. Candidate scenarios are synthetic, owner access is not yet open, and no public signup or guest trial is offered.',
      boundaryLink: 'Why Findur is taking this approach',
    },
    about: {
      eyebrow: 'About Findur',
      titleBefore: 'Compatibility can begin with',
      titleAccent: 'how we choose.',
      intro:
        'Findur is an exploration of portfolio-informed dating: a way to notice patience, curiosity, and decision patterns without turning money into a measure of human worth.',
      thesisLabel: 'The premise',
      thesisTitle: 'Context before appearance.',
      thesisBody:
        'Most dating products lead with a face and compress everything else into a short biography. Findur tests a different sequence: begin with an intentionally limited financial pattern, explain why an introduction appears, and reveal identity only after mutual interest.',
      principlesEyebrow: 'Principles · not promises',
      principlesTitle: 'The boundaries shape the product.',
      principles: [
        {
          title: 'Consent stays specific',
          body: 'Connecting a source is separate from selecting accounts, deriving patterns, and choosing what another person may see.',
        },
        {
          title: 'Compatibility is not worth',
          body: 'No wealth leaderboard, universal financial score, or guarantee can describe a person or predict a relationship.',
        },
        {
          title: 'Disclosure remains legible',
          body: 'People should understand what is private, what is derived, and what may become visible before they make a choice.',
        },
      ],
      statusEyebrow: 'Where the work stands',
      statusTitle: 'A careful evaluation build.',
      statusBody:
        'Findur currently pairs private owner data with clearly labeled synthetic candidates. It is not a public multi-user service. Authentication, connected data, and discovery remain outside this public-site slice.',
      backHome: 'Return home',
    },
    consent: {
      eyebrow: 'Secure connection · staged consent',
      title: 'Connect without sharing your brokerage password.',
      intro: 'SnapTrade hosts authorization. Findur never sees or stores the credentials you use with your brokerage.',
      initialTitle: 'What this step allows',
      initialBody: 'Authorization initially permits only connection status and the minimum masked account inventory needed for a later account-selection step.',
      noDefault: 'No account is included by default.',
      laterTitle: 'What remains off',
      laterBody: 'Balances, positions, activities, signal derivation, profile previews, and Discovery stay unavailable until you explicitly include at least one account in a later step.',
      disclosureTitle: 'Private use and possible disclosure are separate choices',
      disclosureBody: 'Connecting does not publish financial information or reveal it to another person. Later inclusion, private analysis, preview, and disclosure each require their own product boundary.',
      limitsTitle: 'Important limits',
      limits: [
        'Findur cannot trade and does not provide financial advice.',
        'Portfolio information is context—not a judgment of wealth, responsibility, compatibility, or personal worth.',
        'Disconnecting SnapTrade means leaving Findur and deleting your data from app-controlled active storage.',
      ],
      action: 'Continue to SnapTrade',
      unavailable: 'Authorization is not available yet. Callback completion will be enabled in the next release step.',
      back: 'Return home',
      eligibility: 'For eligible adult SnapTrade test users only · evaluation demonstration · no public signup',
    },
    status: {
      metaTitle: 'Build status — Findur',
      eyebrow: 'Deployment diagnostic',
      title: 'Findur build status',
      states: {
        checking: 'Checking exact build…',
        match: 'Exact build verified',
        unavailable: 'API unavailable',
        malformed: 'Build identity malformed',
        stale: 'Build identity stale',
        mismatch: 'Build identity mismatch',
      },
      frontendRevision: 'Frontend revision',
      apiRevision: 'API revision',
      notAvailable: 'not available',
    },
    footer: {
      statement: 'Evidence, not advice. Compatibility, not worth.',
      boundary: '18+ evaluation demonstration · No public signup',
      preferences: 'Display preferences',
    },
  },
  fr: {
    localeName: 'Français',
    skipLink: 'Aller au contenu',
    languageLabel: 'Langue',
    themeLabel: 'Thème',
    themes: { system: 'Système', light: 'Clair', dark: 'Sombre' },
    nav: {
      home: 'Accueil',
      about: 'À propos',
      primary: 'Navigation publique',
    },
    owner: {
      action: 'Accès propriétaire',
      unavailable: 'Consultez les limites de la connexion sécurisée.',
    },
    meta: {
      home: {
        title: 'Findur — Des rencontres axées sur le portefeuille',
        description:
          'Découvrez Findur, un concept d’évaluation 18+ pour des rencontres axées sur le portefeuille et une identité révélée progressivement.',
      },
      about: {
        title: 'À propos de Findur',
        description:
          'Découvrez pourquoi Findur explore le contexte du portefeuille, un partage délibéré et une identité révélée progressivement.',
      },
      connect: {
        title: 'Connexion avec SnapTrade — Findur',
        description: 'Consultez le consentement par étapes avant de connecter SnapTrade à Findur de façon sécurisée.',
      },
    },
    home: {
      eyebrow: 'Des rencontres axées sur le portefeuille',
      titleBefore: 'Trouvez une autre trajectoire dans le',
      titleAccent: 'même ciel.',
      intro:
        'Findur explore la compatibilité par la forme d’un portefeuille avant qu’une photo entre en scène. Les indices ouvrent la conversation; l’intérêt mutuel révèle la personne.',
      demoLabel: 'Démonstration d’évaluation 18+',
      aboutAction: 'Explorer l’idée',
      visualSummary:
        'Quatre idées reliées — fondation, curiosité, patience et perspective — forment une constellation. L’illustration exprime des relations, pas une valeur ni un rendement.',
      nodes: {
        foundation: 'Fondation',
        curiosity: 'Curiosité',
        patience: 'Patience',
        perspective: 'Perspective',
      },
      visualCaption: 'Un portefeuille offre un contexte, pas un score.',
      visualNote: 'Motif illustratif · aucune donnée financière réelle',
      howEyebrow: 'Parcours du concept · 01—03',
      howTitle: 'Plus de contexte. Moins de performance.',
      steps: [
        {
          number: '01',
          title: 'Connecter avec intention',
          body: 'Un portefeuille connecté reste privé par défaut. Le concept commence par un consentement clair et des comptes sélectionnés.',
        },
        {
          number: '02',
          title: 'Choisir ce qui est visible',
          body: 'Le partage est un choix, pas un niveau de statut. Un aperçu montre ce qu’une autre personne pourrait comprendre avant tout enregistrement.',
        },
        {
          number: '03',
          title: 'Révéler l’identité ensemble',
          body: 'La photo reste masquée jusqu’à ce que l’intérêt soit mutuel, afin que la première rencontre porte sur les tendances et les intentions.',
        },
      ],
      boundaryEyebrow: 'Limite actuelle',
      boundaryTitle: 'Une démonstration de produit, pas le lancement public d’un service de rencontres.',
      boundaryBody:
        'La version actuelle sert à évaluer une orientation de produit responsable. Les scénarios de profils sont synthétiques, l’accès propriétaire n’est pas encore ouvert et aucune inscription publique ni démo invitée n’est proposée.',
      boundaryLink: 'Pourquoi Findur adopte cette approche',
    },
    about: {
      eyebrow: 'À propos de Findur',
      titleBefore: 'La compatibilité peut commencer par',
      titleAccent: 'nos choix.',
      intro:
        'Findur explore les rencontres éclairées par le portefeuille : une façon de remarquer la patience, la curiosité et les habitudes de décision sans faire de l’argent une mesure de la valeur humaine.',
      thesisLabel: 'Le principe',
      thesisTitle: 'Le contexte avant l’apparence.',
      thesisBody:
        'La plupart des produits de rencontres commencent par un visage et condensent le reste dans une courte biographie. Findur teste un autre ordre : commencer par un motif financier volontairement limité, expliquer pourquoi une rencontre apparaît et révéler l’identité seulement après un intérêt mutuel.',
      principlesEyebrow: 'Des principes · pas des promesses',
      principlesTitle: 'Les limites façonnent le produit.',
      principles: [
        {
          title: 'Le consentement reste précis',
          body: 'Connecter une source est distinct de la sélection des comptes, de la création de tendances et du choix de ce qu’une autre personne peut voir.',
        },
        {
          title: 'La compatibilité n’est pas la valeur',
          body: 'Aucun classement de richesse, score financier universel ou garantie ne peut décrire une personne ni prédire une relation.',
        },
        {
          title: 'Le partage reste lisible',
          body: 'Chacun devrait comprendre ce qui est privé, ce qui est dérivé et ce qui peut devenir visible avant de faire un choix.',
        },
      ],
      statusEyebrow: 'État d’avancement',
      statusTitle: 'Une version d’évaluation prudente.',
      statusBody:
        'Findur associe actuellement les données privées du propriétaire à des profils synthétiques clairement identifiés. Il ne s’agit pas d’un service public multiutilisateur. L’authentification, les données connectées et la découverte restent hors de cette partie du site public.',
      backHome: 'Retour à l’accueil',
    },
    consent: {
      eyebrow: 'Connexion sécurisée · consentement par étapes',
      title: 'Connectez-vous sans partager votre mot de passe de courtage.',
      intro: 'SnapTrade héberge l’autorisation. Findur ne voit ni ne conserve jamais les identifiants utilisés auprès de votre maison de courtage.',
      initialTitle: 'Ce que cette étape autorise',
      initialBody: 'L’autorisation permet d’abord uniquement de connaître l’état de la connexion et l’inventaire minimal et masqué des comptes nécessaire à une sélection ultérieure.',
      noDefault: 'Aucun compte n’est inclus par défaut.',
      laterTitle: 'Ce qui demeure désactivé',
      laterBody: 'Les soldes, positions, activités, signaux dérivés, aperçus du profil et la découverte demeurent indisponibles jusqu’à ce que vous incluiez explicitement au moins un compte lors d’une étape ultérieure.',
      disclosureTitle: 'L’usage privé et un partage éventuel sont des choix distincts',
      disclosureBody: 'La connexion ne publie aucune information financière et ne la révèle à personne. L’inclusion, l’analyse privée, l’aperçu et le partage nécessitent chacun une décision distincte.',
      limitsTitle: 'Limites importantes',
      limits: [
        'Findur ne peut effectuer aucune opération et ne fournit aucun conseil financier.',
        'Le portefeuille offre un contexte : il ne juge ni la richesse, ni la responsabilité, ni la compatibilité, ni la valeur personnelle.',
        'Déconnecter SnapTrade signifie quitter Findur et supprimer vos données du stockage actif contrôlé par l’application.',
      ],
      action: 'Continuer vers SnapTrade',
      unavailable: 'L’autorisation n’est pas encore disponible. Le retour d’autorisation sera activé à la prochaine étape de livraison.',
      back: 'Retour à l’accueil',
      eligibility: 'Réservé aux adultes admissibles au test SnapTrade · démonstration d’évaluation · aucune inscription publique',
    },
    status: {
      metaTitle: 'État de la version — Findur',
      eyebrow: 'Diagnostic de déploiement',
      title: 'État de la version Findur',
      states: {
        checking: 'Vérification de la version exacte…',
        match: 'Version exacte vérifiée',
        unavailable: 'API indisponible',
        malformed: 'Identité de version non valide',
        stale: 'Identité de version périmée',
        mismatch: 'Identités de version différentes',
      },
      frontendRevision: 'Version de l’interface',
      apiRevision: 'Version de l’API',
      notAvailable: 'non disponible',
    },
    footer: {
      statement: 'Des indices, pas des conseils. Une compatibilité, pas une valeur.',
      boundary: 'Démonstration d’évaluation 18+ · Aucune inscription publique',
      preferences: 'Préférences d’affichage',
    },
  },
} as const

export type Messages = (typeof copy)[Locale]

type I18nContextValue = {
  locale: Locale
  messages: Messages
  setLocale: (locale: Locale) => void
}

const I18nContext = createContext<I18nContextValue | null>(null)

function readLocale(): Locale {
  try {
    return window.localStorage.getItem(localeStorageKey) === 'fr' ? 'fr' : 'en'
  } catch {
    return 'en'
  }
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(readLocale)

  useEffect(() => {
    document.documentElement.lang = locale
  }, [locale])

  const value = useMemo<I18nContextValue>(
    () => ({
      locale,
      messages: copy[locale] as Messages,
      setLocale: (nextLocale) => {
        setLocaleState(nextLocale)
        try {
          window.localStorage.setItem(localeStorageKey, nextLocale)
        } catch {
          // Preferences remain usable for this session when storage is unavailable.
        }
      },
    }),
    [locale],
  )

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}

export function useI18n() {
  const value = useContext(I18nContext)
  if (!value) throw new Error('useI18n must be used within I18nProvider')
  return value
}
