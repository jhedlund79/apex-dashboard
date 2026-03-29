interface ErrorStateProps {
  error: Error | null
  onRetry?: () => void
}

export const ErrorState = ({ error, onRetry }: ErrorStateProps) => (
  <div className="bg-surface border border-danger/20 border-l-[3px] border-l-danger rounded-card p-6 my-2">
    <p className="text-danger text-sm font-semibold mb-1.5">Failed to load data</p>
    <p className="text-content-secondary font-mono text-xs mb-4">
      {error?.message ?? 'Unknown error'}
    </p>
    {onRetry && (
      <button
        onClick={onRetry}
        className="px-4 py-1.5 rounded border border-danger/30 bg-danger/[0.08] text-danger font-mono text-xs transition-colors hover:bg-danger/[0.15]"
      >
        Retry
      </button>
    )}
  </div>
)
