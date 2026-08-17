// PriceBreakdown renders the ordered factor chain that produces a quote's
// final price. It shows the courier *why* a price is what it is — base, time,
// utilization, scarcity — so the dynamic model is transparent.
export default function PriceBreakdown({ quote }) {
  if (!quote) return null
  return (
    <div className="breakdown">
      {quote.breakdown?.map((step, i) => {
        const isFinal = i === quote.breakdown.length - 1
        return (
          <div key={i} className={`step ${isFinal ? 'final' : ''}`}>
            <span>{step.label}</span>
            <span className="factor">
              {step.factor ? `×${step.factor.toFixed(2)} → ` : ''}¥{step.price.toFixed(2)}
            </span>
          </div>
        )
      })}
    </div>
  )
}
