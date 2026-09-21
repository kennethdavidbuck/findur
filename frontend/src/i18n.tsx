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
      initialBody: 'Authorization first lets Findur securely find the accounts you can choose from in the next step.',
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
      checking: 'Checking secure authorization availability…',
      unavailable: 'Authorization is not available yet. Callback completion will be enabled in the next release step.',
      back: 'Return home',
      eligibility: 'For eligible adult SnapTrade test users only · evaluation demonstration · no public signup',
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
          empty: 'We could not find any connected accounts.',
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
          progressSteps: ['1 · Connect', '2 · Choose accounts', '3 · Portfolio showcase', '4 · Preferences & privacy', '5 · Preview & discover'],
          progressNote: 'Next, make your portfolio profile yours.',
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
          reminder: 'You can change these accounts anytime from your Portfolio.',
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
      initialBody: 'L’autorisation permet d’abord à Findur de trouver en toute sécurité les comptes que vous pourrez choisir à l’étape suivante.',
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
      checking: 'Vérification de la disponibilité de l’autorisation sécurisée…',
      unavailable: 'L’autorisation n’est pas encore disponible. Le retour d’autorisation sera activé à la prochaine étape de livraison.',
      back: 'Retour à l’accueil',
      eligibility: 'Réservé aux adultes admissibles au test SnapTrade · démonstration d’évaluation · aucune inscription publique',
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
          empty: 'Nous n’avons trouvé aucun compte connecté.',
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
          progressSteps: ['1 · Connexion', '2 · Choisir les comptes', '3 · Vitrine de portefeuille', '4 · Préférences et confidentialité', '5 · Aperçu et découverte'],
          progressNote: 'Ensuite, personnalisez votre profil de portefeuille.',
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
          reminder: 'Vous pourrez modifier ces comptes en tout temps depuis votre Portefeuille.',
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
