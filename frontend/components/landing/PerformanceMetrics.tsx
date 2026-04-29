export function PerformanceMetrics() {
  const metrics = [
    {
      value: "LiveKit Cloud",
      label: "Powered By",
      detail: "Ultra-low latency global mesh network"
    },
    {
      value: "<300ms",
      label: "Deepgram STT",
      detail: "Near-instant speech to text recognition"
    },
    {
      value: "Local/Edge",
      label: "OpenClaw Inferencing",
      detail: "Bring your own compute and logic"
    },
    {
      value: "Trugen AI",
      label: "Avatar Generation",
      detail: "Real-time lip-sync and expressions"
    }
  ];

  return (
    <section className="py-16 lg:py-24 px-6 border-y border-white/5 bg-gradient-to-b from-transparent to-brand/5">
      <div className="max-w-7xl mx-auto w-full">
        <div className="text-center mb-12">
          <h2 className="text-3xl md:text-4xl font-outfit font-bold text-white mb-4">
            Built for Real-Time Human Interaction
          </h2>
          <p className="text-zinc-400 max-w-2xl mx-auto">
            Our architecture is heavily optimized to keep conversational latency below human perception thresholds seamlessly merging multiple APIs.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          {metrics.map((metric, idx) => (
            <div key={idx} className="text-center p-6 rounded-2xl bg-white/5 border border-white/10 hover:border-brand/50 transition-colors">
              <div className="text-3xl font-bold text-brand mb-2 font-mono">
                {metric.value}
              </div>
              <div className="text-lg font-semibold text-white mb-2 font-outfit">
                {metric.label}
              </div>
              <div className="text-sm text-zinc-500">
                {metric.detail}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
