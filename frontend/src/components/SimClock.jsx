import { useEffect, useState } from 'react'
import { api } from '../api/client.js'

// SimClock shows the simulated wall-clock and lets the user advance it. This
// is what makes time-of-day pricing (e.g. 晚高峰加价) visible during a demo:
// without it the price never changes because real time barely moves between
// page loads. Advancing the clock updates every quote on every page.
export default function SimClock() {
  const [now, setNow] = useState(null)
  const [busy, setBusy] = useState(false)

  // On mount, read the current simulated time without changing it.
  useEffect(() => {
    api.simState().then((st) => setNow(new Date(st.now))).catch(() => {})
  }, [])

  const tick = async (dur) => {
    setBusy(true)
    try {
      const st = await api.simTick(dur)
      setNow(new Date(st.now))
    } finally {
      setBusy(false)
    }
  }

  const reset = async () => {
    setBusy(true)
    try {
      const st = await api.simReset()
      setNow(new Date(st.now))
    } finally {
      setBusy(false)
    }
  }

  const fmt = (d) =>
    d
      ? `${d.toLocaleDateString('zh-CN')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
      : '—'

  const hour = now ? now.getHours() : -1
  const band =
    hour >= 17 && hour < 21
      ? '晚高峰'
      : hour >= 7 && hour < 9
      ? '早高峰'
      : hour >= 22 || hour < 6
      ? '深夜优惠'
      : '平峰'

  return (
    <div className="sim-clock">
      <span className="label">模拟时钟</span>
      <span className="time">{fmt(now)}</span>
      <span className="badge badge-mid">{band}</span>
      <button disabled={busy} onClick={() => tick('1h')} title="推进1小时">
        +1h
      </button>
      <button
        disabled={busy}
        onClick={() => tick('19h')}
        title="跳到约19点晚高峰"
      >
        →晚高峰
      </button>
      <button className="secondary" disabled={busy} onClick={reset}>
        重置
      </button>
    </div>
  )
}
