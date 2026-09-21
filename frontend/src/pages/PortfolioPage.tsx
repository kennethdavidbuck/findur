import { useEffect, useMemo, useRef, useState, type RefObject } from 'react'
import { Button, Dialog, DialogTrigger, Heading, Modal, ModalOverlay } from 'react-aria-components'
import { checkPortfolioInventory, confirmPortfolioInclusion, getPortfolioInclusion, getPortfolioInventory, InventorySessionDefenseError, InventorySessionExpiredError, retryPortfolioInventory, type PortfolioInclusion, type PortfolioInventory } from '../inventory'
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
  const [inclusion, setInclusion] = useState<PortfolioInclusion | null>(null)
  const [draft, setDraft] = useState<Set<string>>(new Set())
  const [failed, setFailed] = useState(false)
  const [inclusionFailed, setInclusionFailed] = useState(false)
  const [retrying, setRetrying] = useState(false)
  const [saving, setSaving] = useState(false)
  const [confirming, setConfirming] = useState(false)
  const [inclusionReload, setInclusionReload] = useState(0)
  const selectAllRef = useRef<HTMLInputElement>(null)
  const confirmButtonRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    let active = true
    void getPortfolioInventory().then((result) => {
      if (active) setInventory(result)
    }).catch((error: unknown) => handleFailure(error, onSessionExpired, () => active && setFailed(true)))
    return () => { active = false }
  }, [onSessionExpired])

  useEffect(() => {
    if (inventory?.state !== 'ready') return
    let active = true
    void getPortfolioInclusion().then((result) => {
      if (!active) return
      setInclusion(result)
      setDraft(recoveryDraft(result))
      setInclusionFailed(false)
    }).catch((error: unknown) => handleFailure(error, onSessionExpired, () => active && setInclusionFailed(true)))
    return () => { active = false }
  }, [inventory?.state, inclusionReload, onSessionExpired])

  const accounts = useMemo(() => inventory?.connections.flatMap((connection) => connection.accounts) ?? [], [inventory])
  const selectable = useMemo(() => accounts.filter((account) => account.selectable), [accounts])
  const labels = useMemo(() => new Map(accounts.map((account) => [account.id, account.maskedLabel])), [accounts])
  const committed = useMemo(() => new Set(inclusion?.committed ?? []), [inclusion])
  const additions = accounts.filter((account) => draft.has(account.id) && !committed.has(account.id))
  const removals = accounts.filter((account) => !draft.has(account.id) && committed.has(account.id))
  const allSelected = selectable.length > 0 && selectable.every((account) => draft.has(account.id))
  const someSelected = selectable.some((account) => draft.has(account.id)) && !allSelected

  useEffect(() => {
    if (selectAllRef.current) selectAllRef.current.indeterminate = someSelected
  }, [someSelected])

  const retry = async () => {
    setRetrying(true)
    setFailed(false)
    try {
      setInventory(await retryPortfolioInventory())
    } catch (error: unknown) {
      handleFailure(error, onSessionExpired, () => setFailed(true))
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
      handleFailure(error, onSessionExpired, () => setFailed(true))
    } finally {
      setRetrying(false)
    }
  }

  const toggleAccount = (id: string) => {
    setDraft((current) => {
      const next = new Set(current)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const save = async () => {
    if (!inclusion) return
    setSaving(true)
    setInclusionFailed(false)
    updateConfirmation(false)
    try {
      const result = await confirmPortfolioInclusion(inclusion.version, [...draft].sort())
      setInclusion(result)
      setDraft(recoveryDraft(result))
    } catch (error: unknown) {
      handleFailure(error, onSessionExpired, () => setInclusionFailed(true))
    } finally {
      setSaving(false)
    }
  }

  function updateConfirmation(open: boolean) {
    setConfirming(open)
    if (!open) requestAnimationFrame(() => confirmButtonRef.current?.focus())
  }

  const toggleAllSelectable = () => {
    setDraft((current) => {
      const next = new Set(current)
      for (const account of selectable) {
        if (allSelected) next.delete(account.id)
        else next.add(account.id)
      }
      return next
    })
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
                      <span>{copy.inclusionReadiness}: {account.selectable ? copy.selectableForInclusion : copy.unavailableForInclusion}</span>
                      <span>{copy.syncState}: {copy.syncStates[account.syncState]}</span>
                      <span>{copy.usabilityReasons[account.usabilityReason]}</span>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          ))}
        </div>
      )}
      {state === 'ready' && (
        <section className="account-inclusion" aria-labelledby="account-inclusion-title">
          <h2 id="account-inclusion-title">{copy.inclusion.title}</h2>
          <p>{copy.inclusion.intro}</p>
          {!inclusion && !inclusionFailed && <p role="status">{copy.inclusion.loading}</p>}
          {inclusionFailed && <div className="inclusion-recovery" role="alert"><p>{copy.inclusion.loadFailed}</p><button className="action action--secondary" type="button" onClick={() => setInclusionReload((value) => value + 1)}>{copy.inclusion.reload}</button></div>}
          {inclusion && (
            <>
              <fieldset className="account-selection" disabled={saving || inclusion.change?.status === 'pending'}>
                <legend>{copy.inclusion.groupName}</legend>
                <label className="select-all">
                  <input ref={selectAllRef} type="checkbox" checked={allSelected} onChange={toggleAllSelectable} />
                  <span>{copy.inclusion.selectAll}</span>
                </label>
                {accounts.map((account) => (
                  <label className={`account-choice${account.selectable ? '' : ' account-choice--disabled'}`} key={account.id}>
                    <input type="checkbox" checked={draft.has(account.id)} disabled={!account.selectable && !committed.has(account.id)} onChange={() => toggleAccount(account.id)} />
                    <span><strong>{account.maskedLabel}</strong><small>{copy.usabilityReasons[account.usabilityReason]}</small></span>
                  </label>
                ))}
              </fieldset>
              <div className="inclusion-summary" aria-live="polite">
                <p>{copy.inclusion.committedCoverage}: {inclusion.committed.length} {copy.inclusion.of} {accounts.length} {copy.inclusion.connectedAccounts}. {copy.inclusion.committed}: {formatLabels(inclusion.committed, labels, copy.inclusion.none)}</p>
                <p>{copy.inclusion.draftCoverage}: {draft.size} {copy.inclusion.of} {accounts.length} {copy.inclusion.connectedAccounts}. {copy.inclusion.draft}: {formatLabels([...draft], labels, copy.inclusion.none)}</p>
                {inclusion.change?.status === 'pending' && <p role="status">{copy.inclusion.pending}</p>}
                {inclusion.change?.status === 'failed' && <p role="alert">{copy.inclusion.failures[inclusion.change.failureReason ?? 'provider_unavailable']}</p>}
                {saving && <p role="status">{copy.inclusion.saving}</p>}
              </div>
              {(inclusion.change?.status === 'pending' || inclusion.change?.status === 'failed') && (
                <div className="inclusion-recovery">
                  <p>{inclusion.change.status === 'pending' ? copy.inclusion.pendingRecovery : copy.inclusion.failedRecovery}</p>
                  <button className="action action--secondary" type="button" disabled={saving} onClick={() => setInclusionReload((value) => value + 1)}>{copy.inclusion.reloadStatus}</button>
                </div>
              )}
              <DialogTrigger isOpen={confirming} onOpenChange={updateConfirmation}>
                <Button ref={confirmButtonRef} className="action action--primary" isDisabled={saving || additions.length + removals.length === 0}>{inclusion.change?.status === 'failed' || inclusion.change?.status === 'pending' ? copy.inclusion.retry : copy.inclusion.review}</Button>
                <ModalOverlay className="dialog-backdrop" isDismissable>
                  <Modal className="confirmation-dialog">
                    <Dialog aria-labelledby="confirm-inclusion-title">
                      <Heading slot="title" id="confirm-inclusion-title">{copy.inclusion.confirmTitle}</Heading>
                      <p>{copy.inclusion.accessBoundary}</p>
                      <p>{copy.inclusion.consequence}</p>
                      <p>{copy.inclusion.coverageBefore}: {committed.size} {copy.inclusion.of} {accounts.length} {copy.inclusion.connectedAccounts}. {copy.inclusion.coverageAfter}: {draft.size} {copy.inclusion.of} {accounts.length} {copy.inclusion.connectedAccounts}.</p>
                      <ChangeList title={copy.inclusion.additions} accounts={additions.map((account) => ({ label: account.maskedLabel, category: copy.categories[account.category] }))} none={copy.inclusion.none} categoryLabel={copy.category} />
                      <ChangeList title={copy.inclusion.removals} accounts={removals.map((account) => ({ label: account.maskedLabel, category: copy.categories[account.category] }))} none={copy.inclusion.none} categoryLabel={copy.category} />
                      <h3>{copy.inclusion.purposesHeading}</h3>
                      <ul>{copy.inclusion.purposes.map((purpose) => <li key={purpose}>{purpose}</li>)}</ul>
                      {removals.length > 0 && <p className="danger-notice"><strong>{copy.inclusion.destructiveLabel}:</strong> {copy.inclusion.purgeConsequence}</p>}
                      <div className="dialog-actions">
                        <Button slot="close" className="action action--secondary" autoFocus>{copy.inclusion.cancel}</Button>
                        <Button className={`action ${removals.length > 0 ? 'action--danger' : 'action--primary'}`} onPress={() => { void save() }}>{removals.length > 0 ? copy.inclusion.confirmRemoval : copy.inclusion.confirm}</Button>
                      </div>
                    </Dialog>
                  </Modal>
                </ModalOverlay>
              </DialogTrigger>
            </>
          )}
        </section>
      )}
      {state === 'pending' && <button className="action action--secondary" type="button" disabled={retrying} onClick={() => { void checkStatus() }}>{retrying ? copy.checking : copy.checkStatus}</button>}
      {recovery === 'reconnect' && <button className="action action--primary" type="button" onClick={onReconnect}>{copy.reconnect}</button>}
      {recovery === 'retry' && <button className="action action--primary" type="button" disabled={retrying} onClick={() => { void retry() }}>{retrying ? copy.retrying : copy.retry}</button>}
      <p className="inventory-boundary">{copy.boundary}</p>
    </section>
  )
}

function ChangeList({ title, accounts, none, categoryLabel }: { title: string, accounts: { label: string, category: string }[], none: string, categoryLabel: string }) {
  return <section><h3>{title}</h3>{accounts.length === 0 ? <p>{none}</p> : <ul>{accounts.map((account) => <li key={account.label}>{account.label} — {categoryLabel}: {account.category}</li>)}</ul>}</section>
}

function recoveryDraft(inclusion: PortfolioInclusion) {
  const next = new Set(inclusion.committed)
  if (inclusion.change?.status === 'failed' || inclusion.change?.status === 'pending') {
    for (const addition of inclusion.change.additions) next.add(addition)
  }
  return next
}

function formatLabels(ids: string[], labels: Map<string, string>, none: string) {
  return ids.length === 0 ? none : ids.map((id) => labels.get(id) ?? id).join(', ')
}

function handleFailure(error: unknown, onSessionExpired: () => void, fallback: () => void) {
  if (error instanceof InventorySessionExpiredError || error instanceof InventorySessionDefenseError) onSessionExpired()
  else fallback()
}
