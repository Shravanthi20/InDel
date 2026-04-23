import React, { useEffect, useRef, useState } from "react";
import { motion, useInView } from "framer-motion";

const metrics = [
  {
    value: 68000,
    prefix: "₹",
    suffix: "",
    label: "Premiums Collected",
    sub: "1,000 workers · Chennai · 1 month",
    color: "text-primary-light",
  },
  {
    value: 44000,
    prefix: "₹",
    suffix: "",
    label: "Payouts Disbursed",
    sub: "Auto-processed via Kafka → UPI",
    color: "text-primary-light",
  },
  {
    value: 35,
    prefix: "",
    suffix: "%",
    label: "Gross Margin",
    sub: "Already profitable before scale",
    color: "text-green-300",
  },
  {
    value: 65,
    prefix: "~",
    suffix: "%",
    label: "Loss Ratio",
    sub: "Industry benchmark: 70–85%",
    color: "text-yellow-300",
  },
];

function Counter({ value, prefix, suffix, color }: { value: number; prefix: string; suffix: string; color: string }) {
  const [count, setCount] = useState(0);
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, { once: true });

  useEffect(() => {
    if (!inView) return;
    let start = 0;
    const duration = 1600;
    const step = value / (duration / 16);
    const timer = setInterval(() => {
      start = Math.min(start + step, value);
      setCount(Math.floor(start));
      if (start >= value) clearInterval(timer);
    }, 16);
    return () => clearInterval(timer);
  }, [inView, value]);

  return (
    <div ref={ref} className={`text-4xl md:text-5xl font-extrabold ${color} font-mono`}>
      {prefix}{count.toLocaleString("en-IN")}{suffix}
    </div>
  );
}

export default function PilotNumbersSection() {
  return (
    <section id="impact" className="py-28 dark-section relative overflow-hidden">
      <div className="absolute inset-0 grid-overlay opacity-20 pointer-events-none" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[400px] bg-primary/8 blur-3xl rounded-full pointer-events-none" />

      <div className="relative z-10 max-w-6xl mx-auto px-6">
        {/* Tag */}
        <motion.div
          className="flex justify-center mb-6"
          initial={{ opacity: 0, y: -12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
        >
          <span className="section-tag">📈 Pilot Results</span>
        </motion.div>

        <motion.h2
          className="text-4xl md:text-5xl font-extrabold text-white text-center mb-4 leading-tight"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.1 }}
        >
          Already inside the
          <br />
          <span className="text-gradient-light">profitable band.</span>
        </motion.h2>
        <motion.p
          className="text-white/50 text-lg text-center max-w-2xl mx-auto mb-16"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.25 }}
        >
          Pilot simulation — 1,000 workers, Chennai, one month. <br className="hidden md:block" />
          Numbers validated by the ML pricing model and loss ratio projections.
        </motion.p>

        {/* Metric counters */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6 mb-16">
          {metrics.map((m, i) => (
            <motion.div
              key={i}
              className="glass-dark rounded-2xl p-6 text-center card-hover"
              initial={{ opacity: 0, y: 28 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.55, delay: i * 0.1 }}
            >
              <Counter value={m.value} prefix={m.prefix} suffix={m.suffix} color={m.color} />
              <p className="text-white font-semibold text-sm mt-2 mb-1">{m.label}</p>
              <p className="text-white/40 text-xs">{m.sub}</p>
            </motion.div>
          ))}
        </div>

        {/* Profit band visual */}
        <motion.div
          className="max-w-3xl mx-auto rounded-2xl glass-dark p-8"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          <p className="text-white/50 text-xs uppercase tracking-widest font-semibold text-center mb-6">
            Loss ratio comparison
          </p>
          <div className="relative h-10 rounded-full bg-white/5 overflow-hidden mb-4">
            {/* Industry range */}
            <div className="absolute top-0 h-full bg-white/10 rounded-full" style={{ left: "70%", width: "15%" }} />
            {/* InDel bar */}
            <motion.div
              className="absolute top-1 bottom-1 rounded-full bg-gradient-to-r from-primary to-cyan-400"
              style={{ left: 0 }}
              initial={{ width: 0 }}
              whileInView={{ width: "65%" }}
              viewport={{ once: true }}
              transition={{ duration: 1.2, delay: 0.4, ease: "easeOut" }}
            />
            {/* InDel label */}
            <div className="absolute right-[35%] top-1/2 -translate-y-1/2 text-xs font-bold text-white whitespace-nowrap">
              InDel ~65%
            </div>
          </div>
          <div className="flex justify-between text-xs text-white/30">
            <span>0%</span>
            <span>Industry range (70–85%)</span>
            <span>100%</span>
          </div>
          <p className="text-white/50 text-xs text-center mt-4 italic">
            InDel is below the industry benchmark — and we haven't reached scale yet.
          </p>
        </motion.div>
      </div>
    </section>
  );
}
