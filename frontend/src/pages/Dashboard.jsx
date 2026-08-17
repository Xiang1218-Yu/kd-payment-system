import { useEffect, useState } from 'react'
import { api } from '../api/client.js'
import UtilizationBar from '../components/UtilizationBar.jsx'

// Dashboard is the overview page: region utilization ranking, top crowded vs
// idle cabinets, hourly pickup volume, and average dwell. It answers "where
// is the load imbalanced right now?".
export default function Dashboard() {
  const [data, setData] = useState(null)
  const [err, setErr] = useState('')

  const load = () => {
    api
      .dashboard()
      .then(setData)
      .catch((e) => setErr(e.message))
  }
  useEffect(load, [])

  if (err) return <div className="card"><p className="error">加载失败：{err}</p></div>
  if (!data) return <div className="card"><p className="empty">加载中…</p></div>

  const totalUtil = avg(data.regions?.map((r) => r.utilization))
  const maxVol = Math.max(1, ...data.hourlyVolume)

  return (
    <>
      <div>
        <h1 className="page-title">运营总览</h1>
        <p className="page-desc">
          各区域格口利用率与取件周转，红色为爆满区域、绿色为空闲可调度区域。
        </p>
      </div>

      <div className="stat-row">
        <Stat k="区域数" v={data.regions?.length || 0} sub="覆盖社区分组" />
        <Stat k="平均利用率" v={`${Math.round(totalUtil * 100)}%`} sub="全柜机格口" />
        <Stat k="近7天取件" v={data.totalPickups} sub="历史取件记录" />
        <Stat k="平均停留" v={`${Math.round(data.avgDwellMinutes)}分钟`} sub="包裹占用格口时长" />
      </div>

      <div className="card">
        <h2>区域利用率排行</h2>
        <table>
          <thead>
            <tr>
              <th>区域</th><th>柜机数</th><th className="num">格口</th>
              <th className="num">已占</th><th>利用率</th>
            </tr>
          </thead>
          <tbody>
            {(data.regions || [])
              .slice()
              .sort((a, b) => b.utilization - a.utilization)
              .map((r) => (
                <tr key={r.id}>
                  <td>
                    <strong>{r.name}</strong>
                    <div className="hint">{r.description}</div>
                  </td>
                  <td>{r.cabinetCount}</td>
                  <td className="num">{r.lockerCount}</td>
                  <td className="num">{r.occupiedCount}</td>
                  <td style={{ minWidth: 140 }}>
                    <UtilizationBar util={r.utilization} showLabel />
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>

      <div className="card-grid" style={{ gridTemplateColumns: '1fr 1fr' }}>
        <div className="card">
          <h2>🔴 最拥挤柜机</h2>
          <CabinetTable rows={data.topCrowded} />
        </div>
        <div className="card">
          <h2>🟢 最空闲柜机</h2>
          <CabinetTable rows={data.topIdle} />
        </div>
      </div>

      <div className="card">
        <h2>近7天取件时段分布</h2>
        <p className="hint">晚高峰（17-21点）取件量最高，正是动态加价要削峰的时段。</p>
        <div style={{ display: 'flex', alignItems: 'flex-end', gap: 3, height: 120, marginTop: 12 }}>
          {data.hourlyVolume.map((v, h) => (
            <div key={h} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4 }}>
              <div
                title={`${h}点: ${v}件`}
                style={{
                  width: '100%',
                  height: `${(v / maxVol) * 100}%`,
                  minHeight: v > 0 ? 2 : 0,
                  background: h >= 17 && h < 21 ? 'var(--danger)' : h >= 7 && h < 9 ? 'var(--warning)' : 'var(--primary)',
                  borderRadius: '3px 3px 0 0',
                }}
              />
              <span className="hint" style={{ fontSize: 9 }}>{h}</span>
            </div>
          ))}
        </div>
      </div>
    </>
  )
}

function Stat({ k, v, sub }) {
  return (
    <div className="stat">
      <div className="k">{k}</div>
      <div className="v">{v}</div>
      {sub && <div className="sub">{sub}</div>}
    </div>
  )
}

function CabinetTable({ rows }) {
  if (!rows?.length) return <p className="empty">暂无数据</p>
  return (
    <table>
      <thead>
        <tr><th>柜机</th><th>区域</th><th className="num">利用率</th></tr>
      </thead>
      <tbody>
        {rows.map((c) => (
          <tr key={c.cabinetId}>
            <td>{c.cabinetName}</td>
            <td className="hint">{c.regionName}</td>
            <td style={{ minWidth: 110 }}>
              <UtilizationBar util={c.utilization} showLabel />
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

function avg(arr) {
  if (!arr?.length) return 0
  return arr.reduce((a, b) => a + (b || 0), 0) / arr.length
}
