import { Routes, Route, NavLink } from 'react-router-dom'
import Dashboard from './pages/Dashboard.jsx'
import Pricing from './pages/Pricing.jsx'
import Dropoff from './pages/Dropoff.jsx'
import Cabinets from './pages/Cabinets.jsx'
import SimClock from './components/SimClock.jsx'

// App is the layout shell: a top bar with navigation and the simulated-clock
// control, plus the routed page below. Each page owns its data fetching.
export default function App() {
  return (
    <div className="app">
      <header className="topbar">
        <div className="brand">
          <span className="logo">📦</span>
          <div>
            <h1>智能快递柜动态定价与调度系统</h1>
            <p className="subtitle">基于历史取件数据与实时库存的格口价格与负载调度</p>
          </div>
        </div>
        <SimClock />
      </header>
      <nav className="nav">
        <NavLink to="/" end>总览看板</NavLink>
        <NavLink to="/pricing">动态定价</NavLink>
        <NavLink to="/dropoff">投递调度</NavLink>
        <NavLink to="/cabinets">柜机状态</NavLink>
      </nav>
      <main className="content">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/pricing" element={<Pricing />} />
          <Route path="/dropoff" element={<Dropoff />} />
          <Route path="/cabinets" element={<Cabinets />} />
        </Routes>
      </main>
      <footer className="footer">
        kd-payment-system · Go + React(JSX) · 单一职责分层架构
      </footer>
    </div>
  )
}
