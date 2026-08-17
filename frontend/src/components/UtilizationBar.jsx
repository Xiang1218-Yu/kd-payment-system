// UtilizationBar renders a horizontal bar colored by load band. One visual
// primitive reused across the dashboard, pricing, and dropoff pages.
export default function UtilizationBar({ util, showLabel = false }) {
  const pct = Math.round((util || 0) * 100)
  const cls = util >= 0.9 ? 'util-high' : util >= 0.8 ? 'util-mid' : 'util-low'
  return (
    <div>
      <div className="util-bar">
        <span className={cls} style={{ width: `${Math.min(100, pct)}%` }} />
      </div>
      {showLabel && <span className="hint">{pct}%</span>}
    </div>
  )
}
