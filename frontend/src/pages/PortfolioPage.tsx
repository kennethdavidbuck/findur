import { useEffect, useMemo, useRef, useState, type RefObject } from 'react'
import { Button, Dialog, DialogTrigger, Heading, Modal, ModalOverlay } from 'react-aria-components'
import { checkPortfolioInventory, confirmPortfolioInclusion, getPortfolioInclusion, getPortfolioInventory, InventorySessionDefenseError, InventorySessionExpiredError, retryPortfolioInventory, type PortfolioInclusion, type PortfolioInventory } from '../inventory'
import { useI18n } from '../i18n'

const maxIncludedAccounts = 5

type Props = {
  editing?: boolean
  initialInclusion?: PortfolioInclusion
  headingRef: RefObject<HTMLHeadingElement | null>
  onComplete: (savedNow: boolean) => void
  onReconnect: () => void
  onSessionExpired: () => void
}

export function PortfolioPage({ editing = false, initialInclusion, headingRef, onComplete, onReconnect, onSessionExpired }: Props) {
  const { locale, messages } = useI18n()
  const copy = messages.authenticated.inventory
  const [inventory, setInventory] = useState<PortfolioInventory | null>(null)
  const [inclusion, setInclusion] = useState<PortfolioInclusion | null>(initialInclusion ?? null)
  const [draft, setDraft] = useState<Set<string>>(() => initialInclusion ? recoveryDraft(initialInclusion) : new Set())
  const [failed, setFailed] = useState(false)
  const [inclusionFailed, setInclusionFailed] = useState(false)
  const [saveFailed, setSaveFailed] = useState(false)
  const [retrying, setRetrying] = useState(false)
  const [saving, setSaving] = useState(false)
  const [confirming, setConfirming] = useState(false)
  const [inclusionReload, setInclusionReload] = useState(0)
  const selectAllRef = useRef<HTMLInputElement>(null)
  const reviewButtonRef = useRef<HTMLButtonElement>(null)
  const activeRef = useRef(true)

  useEffect(() => () => { activeRef.current = false }, [])

  useEffect(() => {
    let active = true
    void getPortfolioInventory().then((result) => {
      if (active) setInventory(result)
    }).catch((error: unknown) => handleFailure(error, onSessionExpired, () => active && setFailed(true)))
    return () => { active = false }
  }, [onSessionExpired])

  useEffect(() => {
    if (!inventory || initialInclusion) return
    let active = true
    void getPortfolioInclusion().then((result) => {
      if (!active) return
      setInclusion(result)
      setDraft(recoveryDraft(result))
      setInclusionFailed(false)
    }).catch((error: unknown) => handleFailure(error, onSessionExpired, () => active && setInclusionFailed(true)))
    return () => { active = false }
  }, [inclusionReload, initialInclusion, inventory, onSessionExpired])

  const committed = useMemo(() => new Set(inclusion?.committed ?? []), [inclusion])
  const visibleConnections = useMemo(() => inventory?.connections.map((connection) => ({
    ...connection,
    accounts: connection.accounts.filter((account) => account.usabilityReason !== 'account_closed' &&
      (!temporaryUsabilityReasons.has(account.usabilityReason) || committed.has(account.id) || draft.has(account.id))),
  })).filter((connection) => connection.accounts.length > 0) ?? [], [committed, draft, inventory])
  const accounts = useMemo(() => {
    const collator = new Intl.Collator(locale === 'fr' ? 'fr-CA' : 'en-CA', { numeric: true, sensitivity: 'base' })
    return visibleConnections
      .flatMap((connection) => connection.accounts.map((account) => ({ ...account, brokerageLabel: connection.brokerageLabel })))
      .sort((left, right) => Number(right.selectable) - Number(left.selectable)
        || collator.compare(left.maskedLabel, right.maskedLabel)
        || left.id.localeCompare(right.id))
  }, [locale, visibleConnections])
  const selectable = useMemo(() => accounts.filter((account) => account.selectable), [accounts])
  const additions = accounts.filter((account) => draft.has(account.id) && !committed.has(account.id))
  const removals = accounts.filter((account) => !draft.has(account.id) && committed.has(account.id))
  const committedVisibleCount = accounts.filter((account) => committed.has(account.id)).length
  const selectedVisibleCount = draftVisibleCount(draft, accounts)
  const selectedSelectableCount = selectable.filter((account) => draft.has(account.id)).length
  const allSelected = selectable.length > 0 && selectedSelectableCount === Math.min(selectable.length, maxIncludedAccounts)
  const someSelected = selectable.some((account) => draft.has(account.id)) && !allSelected
  const selectionLimitReached = draft.size >= maxIncludedAccounts
  const currentChangeDraft = inclusion ? sameIDs([...draft], [...recoveryDraft(inclusion)]) : false

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
    setSaveFailed(false)
    setDraft((current) => {
      const next = new Set(current)
      if (next.has(id)) next.delete(id)
      else if (next.size < maxIncludedAccounts) next.add(id)
      return next
    })
  }

  const save = async () => {
    if (!inclusion) return
    if (additions.length + removals.length === 0) {
      updateConfirmation(false)
      onComplete(false)
      return
    }
    const visibleIDs = new Set(accounts.map((account) => account.id))
    const desired = [...draft].filter((id) => committed.has(id) || visibleIDs.has(id)).sort()
    setSaving(true)
    setSaveFailed(false)
    updateConfirmation(false)
    try {
      const result = await confirmPortfolioInclusion(inclusion.version, desired)
      setInclusion(result)
      setDraft(recoveryDraft(result))
      if ((result.change?.status === 'committed' || result.change?.status === 'pending') && activeRef.current) onComplete(true)
    } catch (error: unknown) {
      if (isSessionError(error)) {
        onSessionExpired()
      } else {
        try {
          const durable = await getPortfolioInclusion()
          const pendingSelectionMatches = durable.change?.status === 'pending'
            && sameIDs([...recoveryDraft(durable)], desired)
          if (sameIDs(durable.committed, desired) || pendingSelectionMatches) {
            if (activeRef.current) onComplete(true)
          } else {
            setInclusion(durable)
            setDraft(recoveryDraft(durable))
            setSaveFailed(durable.change?.status !== 'pending' && durable.change?.status !== 'failed')
          }
        } catch (reconciliationError: unknown) {
          handleFailure(reconciliationError, onSessionExpired, () => setSaveFailed(true))
        }
      }
    } finally {
      setSaving(false)
    }
  }

  function updateConfirmation(open: boolean) {
    setConfirming(open)
    if (!open) requestAnimationFrame(() => reviewButtonRef.current?.focus())
  }

  const toggleAllSelectable = () => {
    setSaveFailed(false)
    setDraft((current) => {
      const next = new Set(current)
      for (const account of selectable) {
        if (allSelected) next.delete(account.id)
        else if (next.size < maxIncludedAccounts) next.add(account.id)
      }
      return next
    })
  }

  const state = failed ? 'unavailable' : inventory?.state ?? 'pending'
  const recovery = state === 'empty' || state === 'disabled' || state === 'unauthorized' ? 'reconnect' : state === 'rate_limited' || state === 'unavailable' || state === 'malformed' ? 'retry' : null
  const connectionComplete = state === 'ready'

  return (
    <div className={`connection-setup-grid${editing ? ' connection-setup-grid--edit' : ''}`}>
      {!editing && <aside className="setup-progress" aria-labelledby="setup-progress-title">
        <h2 className="visually-hidden" id="setup-progress-title">{copy.inclusion.progressTitle}</h2>
        <ol>
          {copy.inclusion.progressSteps.map((step, index) => <li className={connectionComplete && index === 0 ? 'done' : index === (connectionComplete ? 1 : 0) ? 'active' : ''} aria-current={index === (connectionComplete ? 1 : 0) ? 'step' : undefined} key={step}>{step}</li>)}
        </ol>
        <p>{copy.inclusion.progressNote}</p>
      </aside>}
      <section className="portfolio-inventory">
      <p className="eyebrow">{connectionComplete ? copy.inclusion.setupEyebrow : copy.inclusion.recoveryEyebrow}</p>
      <h1 ref={headingRef} tabIndex={-1}>{copy.inclusion.setupTitle}</h1>
      <p className="large-copy">{copy.inclusion.setupIntro}</p>
      <div className="setup-layers" aria-label={copy.inclusion.permissionLayers}>
        {connectionComplete ? <span>{copy.inclusion.oauthLayer}</span> : <strong>{copy.inclusion.oauthLayer}</strong>}<i aria-hidden="true" />
        {connectionComplete ? <strong>{copy.inclusion.accountLayer}</strong> : <span>{copy.inclusion.accountLayer}</span>}<i aria-hidden="true" /><span>{copy.inclusion.disclosureLayer}</span>
      </div>
      {state !== 'ready' && <div className="inventory-status" role="status">
          <h2>{copy.connectionHeading}</h2>
          <p>{copy.states[state]}</p>
          {state === 'rate_limited' && inventory?.retryAt && <p>{copy.retryAfter} <time dateTime={inventory.retryAt}>{new Date(inventory.retryAt).toLocaleString(locale === 'fr' ? 'fr-CA' : 'en-CA')}</time></p>}
        </div>}
      {state !== 'ready' && inventory && inventory.connections.length > 0 && (
        <div className="inventory-connections">
          {inventory.connections.map((connection) => (
            <section className="inventory-connection" key={connection.id}>
              <h3>{connection.brokerageLabel}</h3>
              <p>{copy.connectionStates[connection.status]}</p>
            </section>
          ))}
        </div>
      )}
      {state === 'ready' && (
        <section className={`account-inclusion${saving ? ' account-inclusion--saving' : ''}`} aria-busy={saving}>
          <p className="connection-success"><strong>✓ {copy.inclusion.connected}</strong> {copy.inclusion.connectedNext}</p>
          {!inclusion && !inclusionFailed && <p role="status">{copy.inclusion.loading}</p>}
          {inclusionFailed && <div className="inclusion-recovery" role="alert"><p>{copy.inclusion.loadFailed}</p><button className="action action--secondary" type="button" onClick={() => setInclusionReload((value) => value + 1)}>{copy.inclusion.reload}</button></div>}
          {inclusion && accounts.length === 0 && <div className="inclusion-zero-state">
            <p>{copy.inclusion.noAvailableAccounts}</p>
            <button className="action action--primary" type="button" onClick={onReconnect}>{copy.inclusion.manageAccounts}</button>
          </div>}
          {inclusion && accounts.length > 0 && (
            <>
              <fieldset className="account-selection" disabled={saving}>
                <legend>{copy.inclusion.groupName}</legend>
                <p className="selection-policy">{copy.inclusion.selectionPolicy}</p>
                <div className="selection-guidance">
                  <p>{copy.inclusion.selectionLimit}</p>
                  <p className="selection-count" aria-live="polite">{copy.inclusion.selectedCount(selectedVisibleCount)}</p>
                  {selectionLimitReached && <p className="selection-limit" role="status" aria-live="polite">{copy.inclusion.limitReached}</p>}
                </div>
                <label className="select-all">
                  <input ref={selectAllRef} type="checkbox" checked={allSelected} disabled={saving || selectable.length === 0} onChange={toggleAllSelectable} />
                  <span>{copy.inclusion.selectAll}</span>
                </label>
                {accounts.map((account) => {
                      const selected = draft.has(account.id)
                      const choiceDisabled = (!account.selectable && !committed.has(account.id)) || (!selected && selectionLimitReached)
                      return (
                        <label className={`account-choice${choiceDisabled ? ' account-choice--disabled' : ''}`} key={account.id}>
                          <input type="checkbox" checked={selected} disabled={choiceDisabled} onChange={() => toggleAccount(account.id)} />
                          <span className="account-choice__details">
                            <strong>{account.maskedLabel}</strong>
                            <small>{account.brokerageLabel} · {copy.categories[account.category]}</small>
                            {!account.selectable && <small>{copy.usabilityReasons[account.usabilityReason]}</small>}
                          </span>
                          <span className={`account-choice__state${account.selectable ? '' : ' account-choice__state--unavailable'}`}>{account.selectable ? copy.inclusion.ready : copy.inclusion.unavailable}</span>
                        </label>
                      )
                    })}
              </fieldset>
              <div className="inclusion-summary">
                <p><strong>{copy.inclusion.using} {committedVisibleCount} {copy.inclusion.of} {accounts.length} {copy.inclusion.availableAccounts}.</strong></p>
                {(additions.length > 0 || removals.length > 0) && <p>{copy.inclusion.afterSaving}: {selectedVisibleCount} {copy.inclusion.of} {accounts.length} {copy.inclusion.availableAccounts}.</p>}
              </div>
              <DialogTrigger isOpen={confirming} onOpenChange={updateConfirmation}>
                <div className="inclusion-actions">
                  <Button ref={reviewButtonRef} className="action action--primary" isDisabled={saving || selectedVisibleCount === 0}>{copy.inclusion.review}</Button>
                </div>
                <ModalOverlay className="dialog-backdrop" isDismissable>
                  <Modal className="confirmation-dialog confirmation-dialog--compact">
                    <Dialog aria-labelledby="review-account-choices-title">
                      <Heading slot="title" id="review-account-choices-title">{copy.inclusion.confirmTitle}</Heading>
                      <p>{copy.inclusion.resultingSelection}: <strong>{selectedVisibleCount} {copy.inclusion.of} {accounts.length}</strong></p>
                      {additions.length > 0 && <ChangeList title={copy.inclusion.additions} accounts={additions} />}
                      {removals.length > 0 && <ChangeList title={copy.inclusion.removals} accounts={removals} />}
                      <p>{copy.inclusion.reminder}</p>
                      <div className="dialog-actions">
                        <Button slot="close" className="action action--secondary" autoFocus>{copy.inclusion.back}</Button>
                        <Button className="action action--primary" onPress={() => { void save() }}>{copy.inclusion.save}</Button>
                      </div>
                    </Dialog>
                  </Modal>
                </ModalOverlay>
              </DialogTrigger>
            </>
          )}
          {(saving || saveFailed || currentChangeDraft && (inclusion?.change?.status === 'pending' || inclusion?.change?.status === 'failed')) && (
            <div className="inclusion-feedback" role={saveFailed || inclusion?.change?.status === 'failed' ? 'alert' : 'status'}>
              {saving ? <p>{copy.inclusion.saving}</p>
                : saveFailed ? <p>{copy.inclusion.saveFailed}</p>
                  : inclusion?.change?.status === 'pending' ? <p>{copy.inclusion.pending}</p>
                    : <p>{copy.inclusion.failures[inclusion?.change?.failureReason ?? 'provider_unavailable']}</p>}
            </div>
          )}
        </section>
      )}
      {state === 'pending' && <button className="action action--secondary" type="button" disabled={retrying} onClick={() => { void checkStatus() }}>{retrying ? copy.checking : copy.checkStatus}</button>}
      {recovery === 'reconnect' && <button className="action action--primary" type="button" onClick={onReconnect}>{copy.reconnect}</button>}
      {recovery === 'retry' && <button className="action action--primary" type="button" disabled={retrying} onClick={() => { void retry() }}>{retrying ? copy.retrying : copy.retry}</button>}
      </section>
    </div>
  )
}

function ChangeList({ title, accounts }: { title: string, accounts: { id: string, maskedLabel: string, brokerageLabel: string }[] }) {
  return <section><h3>{title}</h3><ul>{accounts.map((account) => <li key={account.id}>{account.maskedLabel} — {account.brokerageLabel}</li>)}</ul></section>
}

function draftVisibleCount(draft: Set<string>, accounts: { id: string }[]) {
  return accounts.filter((account) => draft.has(account.id)).length
}

function sameIDs(first: string[], second: string[]) {
  const sortedSecond = [...second].sort()
  return first.length === second.length && [...first].sort().every((id, index) => id === sortedSecond[index])
}

function recoveryDraft(inclusion: PortfolioInclusion) {
  const next = new Set(inclusion.committed)
  if (inclusion.change?.status === 'failed' || inclusion.change?.status === 'pending') {
    for (const addition of inclusion.change.additions) next.add(addition)
  }
  return next
}

function handleFailure(error: unknown, onSessionExpired: () => void, fallback: () => void) {
  if (isSessionError(error)) onSessionExpired()
  else fallback()
}

function isSessionError(error: unknown) {
  return error instanceof InventorySessionExpiredError || error instanceof InventorySessionDefenseError
}

const temporaryUsabilityReasons = new Set(['sync_pending', 'connection_unavailable', 'account_unavailable', 'sync_unavailable'])
