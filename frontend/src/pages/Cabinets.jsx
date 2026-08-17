import { useEffect, useState } from 'react'
import { api } from '../api/client.js'
import UtilizationBar from '../components/UtilizationBar.jsx'

// Cabinets page: a filterable list of all cabinets with per-size availability
// and utilization. Clicking a row opens its detail (the full locker grid and
// a pickup demo so the load can be released live).
export default function Cabinets() {
  const [regions, setRegions] = useState([])
  const [region, setRegion] = useState('')
  const [cabinets, setCabinets] = useState([])
  const [detail, setDetail] = useState(null)
  const [busy, setBusy] = useState(false)

  const load = (r) => {
    api.cabinets(r).then(setCabinets)
  }
  useEffect(() => {
    api.regions().then(setRegions)
    load('')
  }, [])

  return (
    <>
      <div>
        <h1 className="page-title">柜机状态</h1>
        <p className="page-desc">
          查看每台柜机的格口占用与可用情况。点击柜机查看格口分布并模拟取件。
        </p>
      </div>

      <div className="card">
        <div className="form-row">
          <div className="field">
            <label>按区域筛选</label>
            <select value={region} onChange={(e) => { setRegion(e.target.value); load(e.target.value) }}>
              <option value="">全部区域</option>
              {regions.map((r) => (
                <option key={r.id} value={r.id}>{r.name}</option>
              ))}
            </select>
          </div>
        </div>
        <table style={{ marginTop: 14 }}>
          <thead>
            <tr>
              <th>柜机</th><th>区域</th><th className="num">总格口</th>
              <th className="num">已占</th><th className="num">可用</th>
              <th>利用率</th><th></th>
            </tr>
          </thead>
          <tbody>
            {cabinets.map((c) => (
              <tr key={c.id}>
                <td>
                  <strong>{c.name}</strong>
                  <div className="hint">{c.address}</div>
                </td>
                <td className="hint">{c.regionId}</td>
                <td className="num">{c.total}</td>
                <td className="num">{c.occupied}</td>
                <td className="num">{c.total - c.occupied}</td>
                <td style={{ minWidth: 120 }}>
                  <UtilizationBar util={c.utilization} showLabel />
                </td>
                <td>
                  <button className="secondary" onClick={() => openDetail(c.id, setDetail, setBusy)}>
                    查看
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {detail && (
        <CabinetDetail detail={detail} onClose={() => setDetail(null)} onPickup={() => openDetail(detail.cabinet.id, setDetail, setBusy)} busy={busy} />
      )}
    </>
  )
}

async function openDetail(id, setDetail, setBusy) {
  setBusy(true)
  try {
    const d = await api.cabinet(id)
    setDetail(d)
  } finally {
    setBusy(false)
  }
}

function CabinetDetail({ detail, onClose, onPickup, busy }) {
  const [picked, setPicked] = useState(null)
  const c = detail.cabinet
  const sizes = ['S', 'M', 'L']

  const pickup = async (lockerId) => {
    setBusy?.(true)
    try {
      const res = await api.pickup(lockerId)
      setPicked(res)
      onPickup() // refresh the grid
    } catch (e) {
      setPicked({ error: e.message })
    } finally {
      setBusy?.(false)
    }
  }

  return (
    <div className="card">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>{c.name} · 格口分布</h2>
        <button className="secondary" onClick={onClose}>关闭</button>
      </div>
      <p className="hint" style={{ marginTop: 6 }}>
        {detail.regionName} · {c.address}
      </p>

      <div className="stat-row" style={{ marginTop: 12 }}>
        {sizes.map((sz) => {
          const st = detail.sizeStats?.[sz]
          if (!st || st.total === 0) return null
          return (
            <div className="stat" key={sz}>
              <div className="k">{sz} 格</div>
              <div className="v">{st.available}/{st.total}</div>
              <div className="sub">可用/总数</div>
            </div>
          )
        })}
      </div>

      <div style={{ display: 'flex', gap: 8, margin: '14px 0 8px', alignItems: 'center' }}>
        <span className="hint">图例：</span>
        <span className="locker legend S">S 空</span>
        <span className="locker legend S occ">S 占</span>
        <span className="locker legend M">M 空</span>
        <span className="locker legend M occ">M 占</span>
        <span className="locker legend L">L 空</span>
        <span className="locker legend L occ">L 占</span>
      </div>

      <div className="lockers">
        {c.lockers?.map((l) => (
          <button
            key={l.id}
            className={`locker ${l.size} ${l.occupied ? 'occ' : ''}`}
            title={`${l.id} · ${l.size} · ${l.occupied ? '已占用' : '空闲'}`}
            disabled={!l.occupied}
            onClick={() => l.occupied && pickup(l.id)}
            style={{
              cursor: l.occupied ? 'pointer' : 'default',
              border: 'none',
              padding: 0,
            }}
          >
            {l.size}
          </button>
        ))}
      </div>

      <p className="hint" style={{ marginTop: 10 }}>
        点击已占用（半透明）的格口可模拟取件，释放格口并写入历史记录。
      </p>

      {picked && (
        <div className="card" style={{ marginTop: 12, background: 'var(--surface-2)' }}>
          {picked.error ? (
            <p className="error">{picked.error}</p>
          ) : (
            <p>
              ✅ 取件成功 · 格口 {picked.lockerId} · 实付 ¥{picked.pricePaid.toFixed(2)} ·
              停留 {Math.round(picked.dwellMinutes)} 分钟
            </p>
          )}
        </div>
      )}
    </div>
  )
}
