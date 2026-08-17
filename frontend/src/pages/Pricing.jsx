import { useEffect, useState } from 'react'
import { api } from '../api/client.js'
import UtilizationBar from '../components/UtilizationBar.jsx'
import PriceBreakdown from '../components/PriceBreakdown.jsx'

// Pricing page: pick a cabinet and a parcel size, see the live quote and its
// full factor breakdown. Advancing the SimClock changes the time-of-day
// multiplier visibly — the whole point of the demo.
const SIZES = [
  { id: 'S', name: '小格' },
  { id: 'M', name: '中格' },
  { id: 'L', name: '大格' },
]

export default function Pricing() {
  const [cabinets, setCabinets] = useState([])
  const [cabinetId, setCabinetId] = useState('')
  const [size, setSize] = useState('M')
  const [quote, setQuote] = useState(null)
  const [err, setErr] = useState('')

  useEffect(() => {
    api.cabinets().then((cs) => {
      setCabinets(cs)
      // default to the most crowded cabinet so the page opens showing a
      // surcharge scenario.
      const busiest = cs.slice().sort((a, b) => b.utilization - a.utilization)[0]
      setCabinetId(busiest?.id || '')
    })
  }, [])

  const fetchQuote = () => {
    if (!cabinetId) return
    setErr('')
    api
      .pricing(cabinetId, size)
      .then(setQuote)
      .catch((e) => {
        setQuote(null)
        setErr(e.message)
      })
  }

  // re-fetch when cabinet or size changes, and also expose a manual refresh.
  useEffect(() => {
    if (cabinetId) fetchQuote()
  }, [cabinetId, size])

  const selected = cabinets.find((c) => c.id === cabinetId)

  return (
    <>
      <div>
        <h1 className="page-title">动态定价查询</h1>
        <p className="page-desc">
          报价 = 基准价 × 时段系数 × 利用率系数 × 稀缺系数。推进右上模拟时钟，可观察晚高峰加价生效。
        </p>
      </div>

      <div className="card">
        <div className="form-row">
          <div className="field" style={{ minWidth: 260 }}>
            <label>选择柜机</label>
            <select value={cabinetId} onChange={(e) => setCabinetId(e.target.value)}>
              {cabinets.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name} · {c.regionId} · 利用率{Math.round(c.utilization * 100)}%
                </option>
              ))}
            </select>
          </div>
          <div className="field">
            <label>包裹尺寸</label>
            <select value={size} onChange={(e) => setSize(e.target.value)}>
              {SIZES.map((s) => (
                <option key={s.id} value={s.id}>{s.name}（{s.id}）</option>
              ))}
            </select>
          </div>
          <button className="secondary" onClick={fetchQuote}>刷新报价</button>
        </div>
      </div>

      {err && <div className="card"><p className="error">{err}</p></div>}

      {selected && (
        <div className="card">
          <h2>{selected.name}</h2>
          <p className="hint">{selected.address} · 当前整体利用率 {Math.round(selected.utilization * 100)}%</p>
          <div style={{ maxWidth: 300, margin: '6px 0 4px' }}>
            <UtilizationBar util={selected.utilization} showLabel />
          </div>
          {quote && (
            <div className="stat-row" style={{ marginTop: 12 }}>
              <div className="stat">
                <div className="k">当前报价</div>
                <div className="v">¥{quote.final.toFixed(2)}</div>
                <div className="sub">{quote.timeOfDayLabel} · {size}号格</div>
              </div>
              <div className="stat">
                <div className="k">时段系数</div>
                <div className="v">×{quote.timeFactor.toFixed(2)}</div>
                <div className="sub">{quote.timeOfDayLabel}</div>
              </div>
              <div className="stat">
                <div className="k">利用率系数</div>
                <div className="v">×{quote.utilizationFactor.toFixed(2)}</div>
                <div className="sub">当前 {Math.round(quote.utilization * 100)}%</div>
              </div>
              <div className="stat">
                <div className="k">剩余格口</div>
                <div className="v">{quote.available}</div>
                <div className="sub">稀缺系数 ×{quote.stockFactor.toFixed(2)}</div>
              </div>
            </div>
          )}
        </div>
      )}

      {quote && (
        <div className="card">
          <h2>报价构成明细</h2>
          <p className="hint">每一项系数如何将基准价推到最终价。</p>
          <PriceBreakdown quote={quote} />
        </div>
      )}
    </>
  )
}
