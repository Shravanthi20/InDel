import { Link } from 'react-router-dom'
import { PageShell, Panel } from './OperationsShared'

const cards = [
  {
    title: 'Synthetic Data',
    description: 'Seed deterministic scenarios and artifacts for full-pipeline demos.',
    to: '/operations/synthetic',
  },
  {
    title: 'Weekly Cycle',
    description: 'Run premium recomputation and inspect idempotent cycle outputs.',
    to: '/operations/weekly-cycle',
  },
  {
    title: 'Payout Ops',
    description: 'Generate claims, queue payouts, and process retries from one surface.',
    to: '/operations/payouts',
  },
  {
    title: 'Reconciliation',
    description: 'Validate processed totals and mismatch counts across date windows.',
    to: '/operations/reconciliation',
  },
]

export default function GodMode() {
  return (
    <PageShell
      eyebrow="Operations"
      title="God Mode"
      description="Central operations console for synthetic generation, premium cycles, payout control, and reconciliation."
    >
      <Panel title="Control Deck" subtitle="Pick an operation module below.">
        <div className="grid gap-4 md:grid-cols-2">
          {cards.map((card) => (
            <Link
              key={card.to}
              to={card.to}
              className="rounded-2xl border border-slate-200 bg-white p-5 transition hover:border-[var(--brand-primary)] hover:bg-[var(--brand-soft)]"
            >
              <h3 className="text-lg font-black tracking-tight text-slate-900 dark:text-white">{card.title}</h3>
              <p className="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-300">{card.description}</p>
            </Link>
          ))}
        </div>
      </Panel>
    </PageShell>
  )
}