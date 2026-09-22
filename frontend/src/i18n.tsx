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
      action: 'Log in',
    },
    meta: {
      home: {
        title: 'Findur — Portfolio-first dating',
        description:
          'A portfolio-first dating experience where investing patterns shape first impressions and mutual matches reveal photos.',
      },
      about: {
        title: 'About Findur',
        description:
          'Learn how Findur uses investing patterns as conversation starters without treating wealth as compatibility.',
      },
      connect: {
        title: 'Log in to Findur',
        description: 'Log in to the private Findur demo securely with SnapTrade.',
      },
    },
    home: {
      eyebrow: 'Portfolio-first dating',
      titleBefore: 'Find a different pattern in the',
      titleAccent: 'same sky.',
      intro:
        'Meet the person behind the portfolio. Findur turns investing patterns into a new kind of first impression, then reveals photos when the feeling is mutual.',
      demoLabel: 'Private 18+ demo',
      loginAction: 'Log in',
      aboutAction: 'How Findur works',
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
      howEyebrow: 'How it works · 01—03',
      howTitle: 'Connect. Discover. Match.',
      steps: [
        {
          number: '01',
          title: 'Connect your portfolio',
          body: 'Log in securely with SnapTrade, then choose which investment accounts Findur may use.',
        },
        {
          number: '02',
          title: 'Meet the portfolio first',
          body: 'Discover people through investing patterns and choose what your own profile may show.',
        },
        {
          number: '03',
          title: 'Match, then reveal',
          body: 'Photos stay hidden until interest is mutual, so every reveal starts with a real match.',
        },
      ],
      boundaryEyebrow: 'For now',
      boundaryTitle: 'A private demo.',
      boundaryBody:
        'Findur keeps each signed-in user’s portfolio private and pairs it only with clearly labeled synthetic profiles. Public signup is not open yet.',
      boundaryLink: 'Learn more about Findur',
    },
    about: {
      eyebrow: 'Why Findur',
      titleBefore: 'There’s more to investing than',
      titleAccent: 'the balance.',
      intro:
        'A portfolio can hint at patience, curiosity, concentration, and an appetite for change. Findur uses those patterns as conversation starters—not as proof of compatibility or worth.',
      thesisLabel: 'What changes',
      thesisTitle: 'The portfolio becomes the first impression.',
      thesisBody:
        'Findur looks at the mix of holdings, diversification, recent activity, and account coverage a person chooses to include. Those signals show where investing styles might align or complement each other. Exact detail stays limited by that person’s disclosure choice, and photos appear only after mutual interest.',
      principlesEyebrow: 'How it stays human',
      principlesTitle: 'The numbers never get the last word.',
      principles: [
        {
          title: 'Patterns, not wealth',
          body: 'Findur notices composition, concentration, diversification, and activity—not who has more money or who is more desirable.',
        },
        {
          title: 'Your portfolio, your boundaries',
          body: 'You choose the accounts Findur may use, preview what your profile could show, and save a disclosure level only when it feels right.',
        },
        {
          title: 'One mutual reveal',
          body: 'Both people make one initial choice from the same portfolio-first experience. Mutual interest reveals the photos—there is no second appearance ranking.',
        },
      ],
      statusEyebrow: 'Available today',
      statusTitle: 'A private demo.',
      statusBody:
        'The private demo keeps every signed-in user’s portfolio isolated and pairs it only with clearly labeled synthetic profiles. It is not yet a public dating service.',
      backHome: 'Return home',
    },
    consent: {
      eyebrow: 'Private demo · 18+',
      title: 'Log in to Findur.',
      intro: 'Continue with SnapTrade to connect your brokerage securely. Findur never receives or stores your brokerage credentials.',
      reassurance: 'No account is included by default, and connecting never publishes your portfolio.',
      summaryTitle: 'How your data works',
      summary: [
        'SnapTrade first shares your brokerage name, connection status, and masked account details so you can choose accounts. Balances, holdings, activity, signal building, profile previews, and Discovery stay off until you explicitly include at least one account.',
        'Confirming selected accounts starts private analysis and profile previews. What a match may see remains a separate choice. Anything you screenshot or share can leave Findur.',
        'Findur is read-only—not advice or a measure of wealth or worth. Disconnecting means leaving Findur and deletes your complete account from app-controlled active storage.',
      ],
      action: 'Continue with SnapTrade',
      checking: 'Checking login availability…',
      unavailable: 'Login is unavailable right now. Try again later.',
      back: 'Return home',
      eligibility: 'Approved SnapTrade test users only · No public signup',
    },
    authenticated: {
      navigation: 'Account navigation',
      discovery: 'Discovery',
      portfolio: 'Portfolio',
      profile: 'Profile',
      checking: 'Checking your secure session…',
      recovering: 'Returning to secure connection…',
      portfolioTitle: 'Your portfolio profile is taking shape.',
      portfolioBody: 'Your account choices are saved. Next, you’ll be able to shape how your portfolio appears in your dating profile.',
      portfolioSaved: 'Account choices saved. Your portfolio profile is ready for the next step.',
      editAccounts: 'Edit included accounts',
      inventory: {
        connectionHeading: 'Account setup',
        states: {
          pending: 'Getting your accounts ready…',
          ready: 'Your accounts are ready.',
          empty: 'No investment accounts are ready to include.',
          disabled: 'Your connection needs attention before we can load your accounts.',
          unauthorized: 'Your connection needs to be renewed before we can load your accounts.',
          rate_limited: 'Your accounts are not ready yet. Check again in a moment.',
          unavailable: 'We couldn’t load your accounts.',
          malformed: 'We couldn’t load your accounts.',
        },
        connectionStates: { active: 'Active', disabled: 'Needs repair', unavailable: 'Unavailable' },
        syncModes: { realtime: 'Real time', delayed: 'Delayed', unknown: 'Unknown' },
        syncStates: { complete: 'Complete', pending: 'Pending', unavailable: 'Unavailable', unknown: 'Unknown' },
        categories: { investment: 'Investment', deposit: 'Deposit', credit: 'Line of credit', unknown: 'Category unavailable' },
        usabilityReasons: {
          ready: 'Ready to choose.',
          provisional_status: 'Ready to choose while account details finish updating.',
          provisional_category: 'Ready to choose while account details finish updating.',
          sync_pending: 'Still getting this account ready.',
          connection_disabled: 'This connection needs attention.',
          connection_unavailable: 'We cannot reach this connection right now.',
          account_closed: 'Unavailable because the account is closed.',
          account_unavailable: 'This account is temporarily unavailable.',
          unsupported_category: 'Unavailable because Findur supports investment accounts only.',
          sync_unavailable: 'We could not update this account.',
        },
        inclusion: {
          setupEyebrow: 'OAuth return · connection success',
          recoveryEyebrow: 'Account setup',
          setupTitle: 'Choose what Findur may use.',
          setupIntro: 'Pick the accounts you’d like to include. You can change this anytime.',
          progressTitle: 'Connection setup',
          progressSteps: ['1 · Connect', '2 · Choose accounts', '3 · Portfolio showcase'],
          progressNote: 'Next, review your private Portfolio Showcase.',
          permissionLayers: 'Connection setup steps',
          oauthLayer: '1 · Connect',
          accountLayer: '2 · Choose accounts',
          disclosureLayer: '3 · Portfolio',
          connected: 'Your accounts are ready.',
          connectedNext: 'Now choose the accounts for your profile.',
          loading: 'Loading your saved choices…',
          loadFailed: 'We could not load your saved choices.',
          reload: 'Try loading again',
          groupName: 'Your accounts',
          selectAll: 'Select all accounts',
          using: 'Saved to your profile:',
          afterSaving: 'Selected after saving',
          of: 'of',
          availableAccounts: 'accounts shown',
          ready: 'Ready',
          unavailable: 'Unavailable',
          noAvailableAccounts: 'No accounts are available to add to your profile right now.',
          manageAccounts: 'Reconnect or manage accounts',
          pending: 'Finishing your update…',
          saving: 'Saving your account choices…',
          review: 'Review my choices',
          confirmTitle: 'Ready to save your choices?',
          resultingSelection: 'Your profile will include',
          additions: 'Adding',
          removals: 'Removing',
          reminder: 'Saving opens your Portfolio Showcase. You can change these accounts anytime.',
          back: 'Back',
          save: 'Save my choices',
          saveFailed: 'We couldn’t save every choice. Review them and try again.',
          failures: {
            authorization_required: 'We couldn’t save every choice. Review them and try again.',
            rate_limited: 'We couldn’t save every choice. Review them and try again.',
            provider_unavailable: 'We couldn’t save every choice. Review them and try again.',
            unusable_data: 'We couldn’t save every choice. Review them and try again.',
            stale_guard: 'Your accounts changed while saving. Review the current choices and try again.',
          },
        },
        checkStatus: 'Check again',
        checking: 'Checking…',
        reconnect: 'Reconnect accounts',
        retry: 'Try again',
        retrying: 'Trying again…',
        retryAfter: 'Try again after:',
      },
      discoveryTitle: 'Discovery is not available yet.',
      discoveryBody: 'Complete the portfolio setup before using Discovery.',
      profileTitle: 'Profile setup comes later.',
      profileBody: 'Complete the portfolio setup before building a profile.',
      logout: 'Log out',
      loggingOut: 'Logging out…',
      retryLogout: 'Retry logout',
      logoutFailed: 'Your session is still active. Please retry.',
      metaDescription: 'Findur — portfolio-first dating.',
    },
    status: {
      metaTitle: 'Build status — Findur',
      eyebrow: 'Deployment diagnostic',
      title: 'Findur build status',
      states: {
        checking: 'Checking deployment health…',
        ready: 'Deployment healthy',
        unavailable: 'API unavailable',
      },
      frontendRevision: 'Frontend revision',
      apiRevision: 'API revision',
      notAvailable: 'not available',
    },
    footer: {
      statement: 'Evidence, not advice. Compatibility, not worth.',
      boundary: 'Private 18+ demo · No public signup',
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
      action: 'Se connecter',
    },
    meta: {
      home: {
        title: 'Findur — Des rencontres axées sur le portefeuille',
        description:
          'Une expérience de rencontres où les habitudes d’investissement créent la première impression et les matchs mutuels dévoilent les photos.',
      },
      about: {
        title: 'À propos de Findur',
        description:
          'Découvrez comment Findur transforme les habitudes d’investissement en sujets de conversation sans confondre richesse et compatibilité.',
      },
      connect: {
        title: 'Connectez-vous à Findur',
        description: 'Connectez-vous à la démo privée de Findur de façon sécurisée avec SnapTrade.',
      },
    },
    home: {
      eyebrow: 'Rencontres axées sur le portefeuille',
      titleBefore: 'Trouvez une autre trajectoire dans le',
      titleAccent: 'même ciel.',
      intro:
        'Rencontrez la personne derrière le portefeuille. Findur transforme les habitudes d’investissement en une nouvelle première impression, puis dévoile les photos lorsque l’intérêt est mutuel.',
      demoLabel: 'Démo privée 18+',
      loginAction: 'Se connecter',
      aboutAction: 'Comment fonctionne Findur',
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
      howEyebrow: 'Comment ça marche · 01—03',
      howTitle: 'Connectez. Découvrez. Matchez.',
      steps: [
        {
          number: '01',
          title: 'Connectez votre portefeuille',
          body: 'Connectez-vous en toute sécurité avec SnapTrade, puis choisissez les comptes que Findur peut utiliser.',
        },
        {
          number: '02',
          title: 'Découvrez d’abord le portefeuille',
          body: 'Rencontrez des personnes par leurs habitudes d’investissement et choisissez ce que votre profil peut montrer.',
        },
        {
          number: '03',
          title: 'Matchez, puis dévoilez',
          body: 'Les photos restent masquées jusqu’à ce que l’intérêt soit mutuel, pour que chaque dévoilement commence par un vrai match.',
        },
      ],
      boundaryEyebrow: 'Pour l’instant',
      boundaryTitle: 'Une démo privée.',
      boundaryBody:
        'Findur garde le portefeuille de chaque personne connectée privé et l’associe uniquement à des profils synthétiques clairement identifiés. L’inscription publique n’est pas encore offerte.',
      boundaryLink: 'En savoir plus sur Findur',
    },
    about: {
      eyebrow: 'Pourquoi Findur',
      titleBefore: 'Investir, c’est bien plus que',
      titleAccent: 'le solde.',
      intro:
        'Un portefeuille peut suggérer la patience, la curiosité, la concentration et le goût du changement. Findur transforme ces tendances en sujets de conversation, jamais en preuve de compatibilité ou de valeur.',
      thesisLabel: 'Ce qui change',
      thesisTitle: 'Le portefeuille devient la première impression.',
      thesisBody:
        'Findur observe la composition des placements, la diversification, l’activité récente et la couverture des comptes qu’une personne choisit d’inclure. Ces signaux montrent où les styles d’investissement peuvent se rejoindre ou se compléter. Le niveau de partage choisi limite les détails exacts, et les photos apparaissent seulement après un intérêt mutuel.',
      principlesEyebrow: 'Pour garder l’humain au centre',
      principlesTitle: 'Les chiffres n’ont jamais le dernier mot.',
      principles: [
        {
          title: 'Des tendances, pas la richesse',
          body: 'Findur remarque la composition, la concentration, la diversification et l’activité, jamais qui a le plus d’argent ni qui est plus désirable.',
        },
        {
          title: 'Votre portefeuille, vos limites',
          body: 'Vous choisissez les comptes que Findur peut utiliser, prévisualisez votre profil et enregistrez un niveau de partage seulement lorsqu’il vous convient.',
        },
        {
          title: 'Un seul dévoilement mutuel',
          body: 'Chaque personne fait un premier choix dans la même expérience axée sur le portefeuille. L’intérêt mutuel dévoile les photos, sans second classement fondé sur l’apparence.',
        },
      ],
      statusEyebrow: 'Disponible aujourd’hui',
      statusTitle: 'Une démo privée.',
      statusBody:
        'La démo privée isole le portefeuille de chaque personne connectée et l’associe uniquement à des profils synthétiques clairement identifiés. Il ne s’agit pas encore d’un service de rencontres public.',
      backHome: 'Retour à l’accueil',
    },
    consent: {
      eyebrow: 'Démo privée · 18+',
      title: 'Connectez-vous à Findur.',
      intro: 'Continuez avec SnapTrade pour connecter votre compte de courtage en toute sécurité. Findur ne reçoit ni ne conserve jamais vos identifiants de courtage.',
      reassurance: 'Aucun compte n’est inclus par défaut et la connexion ne publie jamais votre portefeuille.',
      summaryTitle: 'Comment vos données sont utilisées',
      summary: [
        'SnapTrade transmet d’abord le nom de votre maison de courtage, l’état de la connexion et les renseignements masqués de vos comptes pour vous permettre de choisir. Soldes, placements, activités, création de signaux, aperçus du profil et Découverte restent désactivés jusqu’à ce que vous incluiez explicitement au moins un compte.',
        'La confirmation des comptes choisis lance l’analyse privée et les aperçus du profil. Ce qu’un match peut voir reste un choix distinct. Une capture d’écran ou un partage peut quitter Findur.',
        'Findur est en lecture seule : ni conseil ni mesure de richesse ou de valeur. La déconnexion signifie quitter Findur et supprime votre compte complet du stockage actif contrôlé par l’application.',
      ],
      action: 'Continuer avec SnapTrade',
      checking: 'Vérification de la connexion…',
      unavailable: 'La connexion est indisponible pour le moment. Réessayez plus tard.',
      back: 'Retour à l’accueil',
      eligibility: 'Utilisateurs de test SnapTrade autorisés seulement · Aucune inscription publique',
    },
    authenticated: {
      navigation: 'Navigation du compte',
      discovery: 'Découverte',
      portfolio: 'Portefeuille',
      profile: 'Profil',
      checking: 'Vérification de votre session sécurisée…',
      recovering: 'Retour à la connexion sécurisée…',
      portfolioTitle: 'Votre profil de portefeuille prend forme.',
      portfolioBody: 'Vos choix de comptes sont enregistrés. Vous pourrez ensuite façonner la présentation de votre portefeuille dans votre profil de rencontre.',
      portfolioSaved: 'Choix de comptes enregistrés. Votre profil de portefeuille est prêt pour la prochaine étape.',
      editAccounts: 'Modifier les comptes inclus',
      inventory: {
        connectionHeading: 'Configuration des comptes',
        states: {
          pending: 'Préparation de vos comptes…',
          ready: 'Vos comptes sont prêts.',
          empty: 'Aucun compte de placement n’est prêt à être inclus.',
          disabled: 'Votre connexion nécessite votre attention avant que nous puissions charger vos comptes.',
          unauthorized: 'Votre connexion doit être renouvelée avant que nous puissions charger vos comptes.',
          rate_limited: 'Vos comptes ne sont pas encore prêts. Vérifiez de nouveau dans un instant.',
          unavailable: 'Nous n’avons pas pu charger vos comptes.',
          malformed: 'Nous n’avons pas pu charger vos comptes.',
        },
        connectionStates: { active: 'Active', disabled: 'Réparation requise', unavailable: 'Indisponible' },
        syncModes: { realtime: 'Temps réel', delayed: 'Différé', unknown: 'Inconnu' },
        syncStates: { complete: 'Terminée', pending: 'En attente', unavailable: 'Indisponible', unknown: 'Inconnu' },
        categories: { investment: 'Placement', deposit: 'Dépôt', credit: 'Marge de crédit', unknown: 'Catégorie indisponible' },
        usabilityReasons: {
          ready: 'Prêt à choisir.',
          provisional_status: 'Prêt à choisir pendant la mise à jour des détails du compte.',
          provisional_category: 'Prêt à choisir pendant la mise à jour des détails du compte.',
          sync_pending: 'Nous préparons encore ce compte.',
          connection_disabled: 'Cette connexion nécessite votre attention.',
          connection_unavailable: 'Cette connexion est inaccessible pour le moment.',
          account_closed: 'Indisponible, car le compte est fermé.',
          account_unavailable: 'Ce compte est temporairement indisponible.',
          unsupported_category: 'Indisponible, car Findur prend uniquement en charge les comptes de placement.',
          sync_unavailable: 'Nous n’avons pas pu mettre ce compte à jour.',
        },
        inclusion: {
          setupEyebrow: 'Retour OAuth · connexion réussie',
          recoveryEyebrow: 'Configuration des comptes',
          setupTitle: 'Choisissez ce que Findur peut utiliser.',
          setupIntro: 'Choisissez les comptes que vous souhaitez inclure. Vous pourrez modifier ce choix en tout temps.',
          progressTitle: 'Configuration de la connexion',
          progressSteps: ['1 · Connexion', '2 · Choisir les comptes', '3 · Vitrine de portefeuille'],
          progressNote: 'Ensuite, consultez votre vitrine de portefeuille privée.',
          permissionLayers: 'Étapes de configuration de la connexion',
          oauthLayer: '1 · Connexion',
          accountLayer: '2 · Choix des comptes',
          disclosureLayer: '3 · Portefeuille',
          connected: 'Vos comptes sont prêts.',
          connectedNext: 'Choisissez maintenant les comptes de votre profil.',
          loading: 'Chargement de vos choix enregistrés…',
          loadFailed: 'Nous n’avons pas pu charger vos choix enregistrés.',
          reload: 'Réessayer le chargement',
          groupName: 'Vos comptes',
          selectAll: 'Sélectionner tous les comptes',
          using: 'Enregistrés dans votre profil :',
          afterSaving: 'Sélectionnés après l’enregistrement',
          of: 'sur',
          availableAccounts: 'comptes affichés',
          ready: 'Prêt',
          unavailable: 'Indisponible',
          noAvailableAccounts: 'Aucun compte ne peut être ajouté à votre profil pour le moment.',
          manageAccounts: 'Reconnecter ou gérer les comptes',
          pending: 'Finalisation de votre mise à jour…',
          saving: 'Enregistrement de vos choix de comptes…',
          review: 'Vérifier mes choix',
          confirmTitle: 'Prêt à enregistrer vos choix?',
          resultingSelection: 'Votre profil comprendra',
          additions: 'Ajouts',
          removals: 'Retraits',
          reminder: 'L’enregistrement ouvre votre vitrine de portefeuille. Vous pourrez modifier ces comptes en tout temps.',
          back: 'Retour',
          save: 'Enregistrer mes choix',
          saveFailed: 'Nous n’avons pas pu enregistrer tous vos choix. Vérifiez-les et réessayez.',
          failures: {
            authorization_required: 'Nous n’avons pas pu enregistrer tous vos choix. Vérifiez-les et réessayez.',
            rate_limited: 'Nous n’avons pas pu enregistrer tous vos choix. Vérifiez-les et réessayez.',
            provider_unavailable: 'Nous n’avons pas pu enregistrer tous vos choix. Vérifiez-les et réessayez.',
            unusable_data: 'Nous n’avons pas pu enregistrer tous vos choix. Vérifiez-les et réessayez.',
            stale_guard: 'Vos comptes ont changé pendant l’enregistrement. Vérifiez les choix actuels et réessayez.',
          },
        },
        checkStatus: 'Vérifier de nouveau',
        checking: 'Vérification…',
        reconnect: 'Reconnecter les comptes',
        retry: 'Réessayer',
        retrying: 'Nouvelle tentative…',
        retryAfter: 'Réessayer après :',
      },
      discoveryTitle: 'La découverte n’est pas encore disponible.',
      discoveryBody: 'Terminez la configuration du portefeuille avant d’utiliser la découverte.',
      profileTitle: 'La configuration du profil viendra plus tard.',
      profileBody: 'Terminez la configuration du portefeuille avant de créer un profil.',
      logout: 'Se déconnecter',
      loggingOut: 'Déconnexion…',
      retryLogout: 'Réessayer la déconnexion',
      logoutFailed: 'Votre session est toujours active. Veuillez réessayer.',
      metaDescription: 'Findur — des rencontres axées sur le portefeuille.',
    },
    status: {
      metaTitle: 'État de la version — Findur',
      eyebrow: 'Diagnostic de déploiement',
      title: 'État de la version Findur',
      states: {
        checking: 'Vérification de l’état du déploiement…',
        ready: 'Déploiement sain',
        unavailable: 'API indisponible',
      },
      frontendRevision: 'Version de l’interface',
      apiRevision: 'Version de l’API',
      notAvailable: 'non disponible',
    },
    footer: {
      statement: 'Des indices, pas des conseils. Une compatibilité, pas une valeur.',
      boundary: 'Démo privée 18+ · Aucune inscription publique',
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
    const stored = window.localStorage.getItem(localeStorageKey)
    if (stored === 'en' || stored === 'fr') return stored
  } catch {
    // Browser preference still provides a safe default when storage is unavailable.
  }
  const preferences = window.navigator.languages?.length ? window.navigator.languages : [window.navigator.language]
  for (const preference of preferences) {
    const language = preference.toLowerCase().split('-')[0]
    if (language === 'en' || language === 'fr') return language
  }
  return 'en'
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
