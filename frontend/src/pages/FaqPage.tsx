import type { RefObject } from 'react'
import { useI18n } from '../i18n'

const copy = {
  en: {
    route: 'Portfolio help',
    title: 'Portfolio questions, answered.',
    intro: 'A plain-language guide to when Findur updates your portfolio, what its messages mean, and when you need to act.',
    items: [
      {
        id: 'eligible-accounts',
        question: 'Which accounts can I choose?',
        answer: <>Findur checks eligibility on the server. An account must have an active connection, be open (or have no status reported yet), be an investment account (or not yet classified), and have a completed, available holdings sync. Deposit accounts, lines of credit, closed or unavailable accounts, and accounts on disabled or unavailable connections cannot be selected. A cash-only account may qualify when the provider classifies it as an investment account and the other checks pass.</>,
      },
      {
        id: 'account-selection',
        question: 'Does connecting include my accounts automatically?',
        answer: <>No. Connecting lets Findur list eligible accounts, but none is included automatically. You explicitly choose which accounts to include and can change that choice later. You can include up to five accounts.</>,
      },
      {
        id: 'sync-schedule',
        question: 'How often does Findur update my portfolio?',
        answer: <>After a successful update, Findur ordinarily schedules the next account-list and portfolio update about 24 hours later. “Last sync” is when Findur published the saved information. “Next sync” is the ordinary 24-hour expectation or a scheduled retry time. Work can wait in a queue or for the provider, so the displayed time is not a guaranteed completion time and Findur is not a live view.</>,
      },
      {
        id: 'independent-resources',
        question: 'Why do balances, positions, and activities update separately?',
        answer: <>Balances, positions, and recent activities are independent resources. One can finish while another is still syncing or could not be refreshed; a problem with one does not mean the whole account failed. Findur shows the newest 50 accumulated activities, not your complete transaction history or real-time orders.</>,
      },
      {
        id: 'automatic-retries',
        question: 'What happens when an update has a problem?',
        answer: <>Findur retries temporary problems automatically after 1, 2, 4, 8, 16, and 32 minutes, with later retries remaining 32 minutes apart. If the provider asks Findur to wait longer, that later time is used. Values from the last successful sync remain visible when they are still safe to show; otherwise, Findur hides them until newer information is available.</>,
      },
      {
        id: 'user-actions',
        question: 'When do I need to do something?',
        answer: <>Usually, you can wait while Findur syncs or retries. “Check again” reloads the status already saved by Findur, and “Try again” retries the Findur page request; neither forces the provider to synchronize. Reconnect only when Portfolio asks you to renew authorization or repair the connection.</>,
      },
    ],
  },
  fr: {
    route: 'Aide sur le portefeuille',
    title: 'Vos questions sur le portefeuille.',
    intro: 'Un guide en langage simple sur le moment où Findur met votre portefeuille à jour, la signification des messages et les situations qui exigent votre intervention.',
    items: [
      {
        id: 'eligible-accounts',
        question: 'Quels comptes puis-je choisir?',
        answer: <>Findur vérifie l’admissibilité sur le serveur. Un compte doit avoir une connexion active, être ouvert (ou ne pas encore avoir d’état déclaré), être un compte de placement (ou ne pas encore être classé) et avoir une synchronisation des avoirs terminée et disponible. Les comptes de dépôt, les marges de crédit, les comptes fermés ou indisponibles et les comptes liés à une connexion désactivée ou indisponible ne peuvent pas être sélectionnés. Un compte de placement composé uniquement d’encaisse peut être admissible si les autres vérifications réussissent.</>,
      },
      {
        id: 'account-selection',
        question: 'La connexion inclut-elle mes comptes automatiquement?',
        answer: <>Non. La connexion permet à Findur d’afficher les comptes admissibles, mais aucun compte n’est inclus automatiquement. Vous choisissez explicitement les comptes à inclure et pouvez modifier ce choix plus tard. Vous pouvez inclure jusqu’à cinq comptes.</>,
      },
      {
        id: 'sync-schedule',
        question: 'À quelle fréquence Findur met-il mon portefeuille à jour?',
        answer: <>Après une mise à jour réussie, Findur prévoit normalement la prochaine mise à jour de la liste de comptes et du portefeuille environ 24 heures plus tard. « Dernière synchronisation » indique quand Findur a publié les renseignements enregistrés. « Prochaine synchronisation » indique l’heure normalement prévue après 24 heures ou celle d’une nouvelle tentative. Le travail peut attendre dans une file ou dépendre du fournisseur : l’heure affichée ne garantit donc pas l’heure d’achèvement et Findur n’est pas une vue en temps réel.</>,
      },
      {
        id: 'independent-resources',
        question: 'Pourquoi les soldes, les positions et les activités sont-ils mis à jour séparément?',
        answer: <>Les soldes, les positions et les activités récentes sont des ressources indépendantes. Une ressource peut être terminée pendant qu’une autre se synchronise ou n’a pas pu être actualisée; un problème avec une ressource ne signifie pas que tout le compte a échoué. Findur affiche les 50 activités cumulées les plus récentes, et non l’historique complet de vos opérations ni des ordres en temps réel.</>,
      },
      {
        id: 'automatic-retries',
        question: 'Que se passe-t-il lorsqu’une mise à jour échoue?',
        answer: <>Findur réessaie automatiquement après 1, 2, 4, 8, 16 et 32 minutes; les tentatives suivantes restent espacées de 32 minutes. Si le fournisseur demande à Findur d’attendre plus longtemps, cette heure ultérieure est utilisée. Les valeurs de la dernière synchronisation réussie restent visibles lorsqu’elles peuvent encore être affichées sans risque; sinon, Findur les masque jusqu’à ce que de nouveaux renseignements soient disponibles.</>,
      },
      {
        id: 'user-actions',
        question: 'Quand dois-je intervenir?',
        answer: <>Habituellement, vous pouvez attendre pendant que Findur synchronise les données ou réessaie. « Vérifier à nouveau » recharge l’état déjà enregistré par Findur et « Réessayer » relance la demande de la page Findur; aucune de ces actions ne force la synchronisation du fournisseur. Reconnectez-vous seulement si Portefeuille vous demande de renouveler l’autorisation ou de réparer la connexion.</>,
      },
    ],
  },
} as const

export function FaqPage({ headingRef }: { headingRef: RefObject<HTMLHeadingElement | null> }) {
  const { locale } = useI18n()
  const text = copy[locale]

  return <section className="faq-page">
    <header className="faq-intro">
      <p className="profile-route"><span aria-hidden="true">04</span>{text.route}</p>
      <h1 ref={headingRef} tabIndex={-1}>{text.title}</h1>
      <p>{text.intro}</p>
    </header>
    <div className="faq-list">
      {text.items.map((item, index) => <details key={item.id}>
        <summary>
          <span aria-hidden="true">{String(index + 1).padStart(2, '0')}</span>
          <h2>{item.question}</h2>
          <span className="faq-disclosure-indicator" aria-hidden="true"><span>+</span><span>−</span></span>
        </summary>
        <div className="faq-answer"><p>{item.answer}</p></div>
      </details>)}
    </div>
  </section>
}
