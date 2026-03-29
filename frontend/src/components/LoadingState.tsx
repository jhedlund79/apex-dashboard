export const LoadingState = () => (
  <div className="grid grid-cols-3 gap-4 pt-1">
    {Array.from({ length: 6 }).map((_, i) => (
      <div
        key={i}
        className="h-24 rounded-card bg-surface border border-white/[0.04] animate-shimmer"
        style={{ animationDelay: `${i * 0.1}s` }}
      />
    ))}
  </div>
)
