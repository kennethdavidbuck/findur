import { useEffect, useState, type RefObject } from 'react'
import { checkPortfolioInventory, getPortfolioInventory, InventorySessionDefenseError, InventorySessionExpiredError, retryPortfolioInventory, type PortfolioInventory } from '../inventory'
import { useI18n } from '../i18n'

type Props = {
  headingRef: RefObject<HTMLHeadingElement | null>
  onReconnect: () => void
  onSessionExpired: () => void
}

export function PortfolioPage({ headingRef, onReconnect, onSessionExpired }: Props) {
  const { locale, messages } = useI18n()
  const copy = messages.authenticated.inventory
  const [inventory, setInventory] = useState<PortfolioInventory | null>(null)
  const [failed, setFailed] = useState(false)
  const [retrying, setRetrying] = useState(false)

  useEffect(() => {
    let active = true
    void getPortfolioInventory().then((result) => {
      if (active) setInventory(result)
    }).catch((error: unknown) => {
      if (!active) return
      if (error instanceof InventorySessionExpiredError || error instanceof InventorySessionDefenseError) onSessionExpired()
      else setFailed(true)
    })
    return () => { active = false }
  }, [onSessionExpired])

  const retry = async () => {
    setRetrying(true)
    setFailed(false)
    try {
      setInventory(await retryPortfolioInventory())
    } catch (error: unknown) {
      if (error instanceof InventorySessionExpiredError || error instanceof InventorySessionDefenseError) onSessionExpired()
      else setFailed(true)
    } finally {
      setRetrying(false)
    }
  }

  const checkStatus = async () => {
    setRetrying(true)
    setFailed(false)
    try {
      setInventory(await checkPortfolioInventory())
    } catch (error: unknown) {
      if (error instanceof InventorySessionExpiredError || error instanceof InventorySessionDefenseError) onSessionExpired()
      else setFailed(true)
    } finally {
      setRetrying(false)
    }
  }

  const state = failed ? 'unavailable' : inventory?.state ?? 'pending'
  const recovery = state === 'disabled' || state === 'unauthorized' ? 'reconnect' : state === 'rate_limited' || state === 'unavailable' || state === 'malformed' ? 'retry' : null

  return (
    <section className="portfolio-inventory">
      <p className="eyebrow">{messages.authenticated.privateEyebrow}</p>
      <h1 ref={headingRef} tabIndex={-1}>{copy.title}</h1>
      <p className="large-copy">{copy.noDefault}</p>
      <div className="inventory-status" role="status" aria-live="polite">
        <h2>{copy.connectionHeading}</h2>
        <p>{copy.states[state]}</p>
        {state === 'rate_limited' && inventory?.retryAt && <p>{copy.retryAfter} <time dateTime={inventory.retryAt}>{new Date(inventory.retryAt).toLocaleString(locale === 'fr' ? 'fr-CA' : 'en-CA')}</time></p>}
      </div>
      {inventory && inventory.connections.length > 0 && (
        <div className="inventory-connections">
          {inventory.connections.map((connection) => (
            <section className="inventory-connection" key={connection.id}>
              <h3>{connection.brokerageLabel}</h3>
              <p className="mono-label">{copy.connectionStatus}: {copy.connectionStates[connection.status]}</p>
              <p>{copy.availability}: {connection.available ? copy.available : copy.unavailable}</p>
              <p>{copy.eligibility}: {connection.eligible ? copy.eligible : copy.ineligible}</p>
              <p>{copy.syncMode}: {copy.syncModes[connection.syncMode]}</p>
              <h4>{copy.accountsHeading}</h4>
              {connection.accounts.length === 0 ? <p>{copy.noAccounts}</p> : (
                <ul>
                  {connection.accounts.map((account) => (
                    <li key={account.id}>
                      <strong>{account.maskedLabel}</strong>
                      <span>{copy.category}: {copy.categories[account.category]}</span>
                      <span>{copy.accountType}: {account.type}</span>
                      <span>{copy.availability}: {account.available ? copy.available : copy.unavailable}</span>
                      <span>{copy.eligibility}: {account.eligible ? copy.eligible : copy.ineligible}</span>
                      <span>{copy.syncState}: {copy.syncStates[account.syncState]}</span>
                      {!account.available && <span>{copy.accountUnavailable}</span>}
                    </li>
                  ))}
                </ul>
              )}
            </section>
          ))}
        </div>
      )}
      {state === 'pending' && <button className="action action--secondary" type="button" disabled={retrying} onClick={() => { void checkStatus() }}>{retrying ? copy.checking : copy.checkStatus}</button>}
      {recovery === 'reconnect' && <button className="action action--primary" type="button" onClick={onReconnect}>{copy.reconnect}</button>}
      {recovery === 'retry' && <button className="action action--primary" type="button" disabled={retrying} onClick={() => { void retry() }}>{retrying ? copy.retrying : copy.retry}</button>}
      <p className="inventory-boundary">{copy.boundary}</p>
    </section>
  )
}
