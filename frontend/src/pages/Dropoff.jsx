import { useEffect, useState } from 'react'
import { api } from '../api/client.js'
import UtilizationBar from '../components/UtilizationBar.jsx'
import PriceBreakdown from '../components/PriceBreakdown.jsx'

// Dropoff page simulates a courier trying to drop a parcel. It calls the
// scheduling endpoint and shows whether the system confirmed the requested
// cabinet or redirected to a less-loaded neighbor, plus the ranked
// alternatives. On confirm, it actually occupies a locker (the system mutates
// state), so repeated drops move the load.
const SIZES = [
  { id: 'S', name: '小格' },
  { id: 'M', name: '中格' },
  { id: 'L', name: '大格' },
]

export default function Dropoff() {
  const [cabinets, setCabinets] = useState([])
  const [cabinetId, setCabinetId] = useState('')
  const [size, setSize] = useState('M')
  const [result, setResult] = useState(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')

  useEffect(() => {
    api.cabinets().then((cs) => {
      setCabinets(cs)
      // default to a crowded cabinet so the first dropoff demonstrates
      // redirection.
      const busiest = cs.slice().sort((a, b) => b.utilization - a.utilization)[0]
      setCabinetId(busiest?.id || '')
    })
  }, [])

  const submit = () => {
    if (!cabinetId) return
    setBusy(true)
    setErr('')
    api
      .dropoff(cabinetId, size)
      .then((r) => setResult(r))
      .catch((e) => {
        setResult(null)
        setErr(e.message)
      })
      .finally(() => setBusy(false))
  }

  const s = result?.schedule
  const recommendedAlt =
    s?.alternatives?.find((a) => a.cabinetId === s.recommendedCabinetId) ||
    null

  return (
    <>
      <div>
        <h1 className="page-title">投递负载调度</h1>
        <p className="page-desc">
          当目标柜机接近满载，系统自动将包裹调度至利用率更低的邻近柜机，平衡区域负载。
        </p>
      </div>

      <div className="card">
        <div className="form-row">
          <div className="field" style={{ minWidth: 280 }}>
            <label>快递员想投递到的柜机</label>
            <select value={cabinetId} onChange={(e) => setCabinetId(e.target.value)}>
              {cabinets.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name} · 利用率{Math.round(c.utilization * 100)}%
                </option>
              ))}
            </select>
          </div>
          <div className="field">
            <label>包裹尺寸</label>
            <select value={size} onChange={(e) => setSize(e.target.value)}>
              {SIZES.map((sz) => (
                <option key={sz.id} value={sz.id}>{sz.name}（{sz.id}）</option>
              ))}
            </select>
          </div>
          <button disabled={busy} onClick={submit}>
            {busy ? '调度中…' : '发起投递'}
          </button>
        </div>
      </div>

      {err && <div className="card"><p className="error">{err}</p></div>}

      {s && (
        <div className="card">
          <h2>
            调度结果
            {s.isRedirected ? (
              <span className="badge badge-redirect" style={{ marginLeft: 8 }}>已重定向</span>
            ) : (
              <span className="badge badge-ok" style={{ marginLeft: 8 }}>直接投递</span>
            )}
          </h2>
          <p>{s.reason}</p>
          {result?.occupied ? (
            <div className="stat-row" style={{ marginTop: 10 }}>
              <div className="stat">
                <div className="k">已占格口</div>
                <div className="v">{result.lockerId}</div>
                <div className="sub">包裹号 {result.parcelId}</div>
              </div>
              <div className="stat">
                <div className="k">实付价格</div>
                <div className="v">¥{result.pricePaid.toFixed(2)}</div>
                <div className="sub">{size}号格</div>
              </div>
              {s.isRedirected && (
                <div className="stat">
                  <div className="k">调度距离</div>
                  <div className="v">{Math.round(s.distanceMeters)}m</div>
                  <div className="sub">原柜机 → 推荐柜机</div>
                </div>
              )}
            </div>
          ) : (
            <p className="empty">未能占用格口（推荐柜机暂无空闲），请选择下方替代柜机或稍后再试。</p>
          )}

          {s.recommendedQuote && (
            <>
              <h2 style={{ marginTop: 18 }}>推荐柜机报价</h2>
              <PriceBreakdown quote={s.recommendedQuote} />
            </>
          )}
        </div>
      )}

      {s?.alternatives?.length > 0 && (
        <div className="card">
          <h2>邻近可调度柜机（按利用率、距离排序）</h2>
          {s.alternatives.map((alt) => (
            <div
              key={alt.cabinetId}
              className={`alt ${alt.cabinetId === s.recommendedCabinetId && s.isRedirected ? 'recommended' : ''}`}
            >
              <h3>
                {alt.cabinetName}
                {alt.cabinetId === s.recommendedCabinetId && s.isRedirected && (
                  <span className="badge badge-low">推荐</span>
                )}
              </h3>
              <div className="meta">
                距离 {Math.round(alt.distanceMeters)}m · 利用率 {Math.round(alt.utilization * 100)}% ·
                可用 {alt.available} 格 · 报价 ¥{alt.quote?.final.toFixed(2)}
              </div>
              <div style={{ maxWidth: 260, marginTop: 6 }}>
                <UtilizationBar util={alt.utilization} showLabel />
              </div>
            </div>
          ))}
        </div>
      )}
    </>
  )
}
