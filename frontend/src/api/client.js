// API client — single responsibility: talk to the backend. Every page calls
// these functions instead of fetch directly, so the network layer is in one
// place and easy to swap (e.g. for tests or a different transport).

const base = '/api'

async function json(path, opts) {
  const res = await fetch(base + path, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  })
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const body = await res.json()
      msg = body.error || msg
    } catch (_) {
      // response wasn't JSON; keep the status message
    }
    throw new Error(msg)
  }
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  regions: () => json('/regions'),
  cabinets: (region) => json('/cabinets' + (region ? `?region=${region}` : '')),
  cabinet: (id) => json(`/cabinets/${id}`),
  pricing: (cabinetId, size) =>
    json(`/pricing/${cabinetId}?size=${size || 'M'}`),
  dropoff: (cabinetId, size) =>
    json('/dropoff', {
      method: 'POST',
      body: JSON.stringify({ cabinetId, size }),
    }),
  pickup: (lockerId) =>
    json('/pickup', {
      method: 'POST',
      body: JSON.stringify({ lockerId }),
    }),
  dashboard: () => json('/stats/dashboard'),
  simState: () => json('/sim/state'),
  simTick: (duration) =>
    json('/sim/tick', {
      method: 'POST',
      body: JSON.stringify({ duration: duration || '1h' }),
    }),
  simReset: () =>
    json('/sim/reset', { method: 'POST', body: '{}' }),
}
