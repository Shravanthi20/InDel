import React, { useEffect, useRef, useState } from "react";
import { motion, useInView } from "framer-motion";
import { IconChart } from "./AllSectionIcons";

const zones = [
  {
    city: "Tambaram, Chennai",
    risk: "High",
    riskColor: "text-brand-dark",
    riskBg: "bg-brand-soft border-brand-primary/20",
    premium: "₹22",
    payout: "₹800",
    note: "Monsoon + heat zone",
    barWidth: "90%",
    barColor: "bg-brand-dark shadow-[0_0_12px_rgba(190,24,93,0.3)]",
  },
  {
    city: "Rohini, Delhi",
    risk: "Medium",
    riskColor: "text-brand-primary",
    riskBg: "bg-brand-soft/40 border-brand-primary/10",
    premium: "₹17",
    payout: "₹700",
    note: "Seasonal fog + AQI",
    barWidth: "60%",
    barColor: "bg-brand-primary",
  },
  {
    city: "Koramangala, Bengaluru",
    risk: "Low",
    riskColor: "text-brand-primary/60",
    riskBg: "bg-gray-50 border-gray-100",
    premium: "₹12",
    payout: "₹600",
    note: "Stable climate zone",
    barWidth: "30%",
    barColor: "bg-brand-primary/40",
  },
];

function AnimatedNumber({ target, prefix = "", suffix = "" }: { target: number; prefix?: string; suffix?: string }) {
  const [count, setCount] = useState(0);
  const ref = useRef<HTMLSpanElement>(null);
  const inView = useInView(ref, { once: true });

  useEffect(() => {
    if (!inView) return;
    let start = 0;
    const duration = 1400;
    const step = target / (duration / 16);
    const timer = setInterval(() => {
      start = Math.min(start + step, target);
      setCount(Math.floor(start));
      if (start >= target) clearInterval(timer);
    }, 16);
    return () => clearInterval(timer);
  }, [inView, target]);

  return <span ref={ref}>{prefix}{count.toLocaleString("en-IN")}{suffix}</span>;
}

export default function PricingIntelSection() {
  return (
    <section id="pricing" className="py-28 bg-background relative overflow-hidden">
      <div className="absolute -bottom-32 -right-32 w-80 h-80 rounded-full bg-primary-light blur-3xl pointer-events-none" />

      <div className="relative z-10 max-w-6xl mx-auto px-6">
        {/* Tag */}
        <motion.div
          className="flex justify-center mb-6"
          initial={{ opacity: 0, y: -12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
        >
          <span className="section-tag bg-primary/10 text-primary border border-primary/20">
            <IconChart /> Dynamic Pricing Intelligence
          </span>
        </motion.div>

        <motion.h2
          className="text-4xl md:text-5xl font-extrabold text-center text-text-primary mb-4 leading-tight"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.1 }}
        >
          No flat rates.
          <br />
          <span className="text-gradient">Every premium is computed fresh.</span>
        </motion.h2>
        <motion.p
          className="text-text-secondary text-lg text-center max-w-2xl mx-auto mb-16"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.25 }}
        >
          XGBoost trained on 18 features. Risk score recalculated monthly. Every premium is fully auditable via SHAP TreeExplainer.
        </motion.p>

        <div className="grid lg:grid-cols-2 gap-12 items-start">
          {/* Left: Formula */}
          <motion.div
            initial={{ opacity: 0, x: -24 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.65, delay: 0.1 }}
          >
            {/* Risk Score Formula */}
            <div className="rounded-2xl bg-white border border-primary/10 p-6 mb-6 shadow-sm">
              <p className="text-xs uppercase tracking-widest text-text-secondary font-semibold mb-4">Risk Score Formula</p>
              <div className="font-mono text-sm space-y-1.5 text-text-primary">
                <div className="text-base font-bold text-primary">R = </div>
                {[
                  { label: "Order Volatility  ", weight: "×0.24", color: "text-brand-primary/60" },
                  { label: "Earnings Volatility", weight: "×0.22", color: "text-brand-primary/70" },
                  { label: "Disruption Rate    ", weight: "×0.20", color: "text-brand-primary/80" },
                  { label: "Weather Signal     ", weight: "×0.34", color: "text-brand-primary", bold: true },
                ].map((v, i) => (
                  <div key={i} className="flex items-center gap-3 pl-4">
                    <span className={`${v.color} ${v.bold ? "font-bold" : ""}`}>{v.label}</span>
                    <span className={`${v.color} font-semibold`}>{v.weight}</span>
                    {i < 3 && <span className="text-text-secondary ml-auto">+</span>}
                  </div>
                ))}
              </div>
              <p className="mt-4 text-xs text-text-secondary italic">
                Weather leads at 34% — strongest predictor of delivery income loss across Indian urban zones.
              </p>
            </div>

            {/* Premium Formula */}
            <div className="rounded-2xl bg-white border border-primary/10 p-6 shadow-sm">
              <p className="text-xs uppercase tracking-widest text-text-secondary font-semibold mb-4">Premium Formula</p>
              <div className="font-mono text-sm text-text-primary bg-primary-light/40 rounded-xl p-4">
                P = (E<sub>avg</sub> × 0.0375) × (0.72 + R) × VF
              </div>
              <div className="mt-4 space-y-2 text-xs text-text-secondary">
                <div className="flex gap-2"><span className="font-semibold text-text-primary min-w-[80px]">E_avg →</span> 4-week average daily earnings (InDel first-party data)</div>
                <div className="flex gap-2"><span className="font-semibold text-text-primary min-w-[80px]">R →</span> Risk score above, recalculated monthly</div>
                <div className="flex gap-2"><span className="font-semibold text-text-primary min-w-[80px]">VF →</span> Vehicle Factor: 1.04 for EVs · 1.08 for ICE (rewards sustainable delivery)</div>
              </div>
            </div>
          </motion.div>

          {/* Right: Zone cards */}
          <motion.div
            initial={{ opacity: 0, x: 24 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.65, delay: 0.2 }}
            className="space-y-4"
          >
            <p className="text-xs uppercase tracking-widest text-text-secondary font-semibold mb-6">
              Same worker, different zones, different premiums
            </p>
            {zones.map((zone, i) => (
              <div
                key={i}
                className={`rounded-2xl border ${zone.riskBg} p-5 card-hover`}
              >
                <div className="flex items-center justify-between mb-3">
                  <div>
                    <p className="font-bold text-text-primary text-sm">{zone.city}</p>
                    <p className="text-text-secondary text-xs mt-0.5">{zone.note}</p>
                  </div>
                  <span className={`text-xs font-bold px-3 py-1 rounded-full border ${zone.riskBg} ${zone.riskColor}`}>
                    {zone.risk} Risk
                  </span>
                </div>
                {/* Risk bar */}
                <div className="h-1.5 bg-black/5 rounded-full mb-3 overflow-hidden">
                  <div className={`h-full ${zone.barColor} rounded-full`} style={{ width: zone.barWidth }} />
                </div>
                <div className="flex gap-6 text-sm">
                  <div>
                    <span className="text-text-secondary text-xs block">Weekly Premium</span>
                    <span className="font-bold text-text-primary">{zone.premium}</span>
                  </div>
                  <div>
                    <span className="text-text-secondary text-xs block">Max Weekly Payout</span>
                    <span className="font-bold text-text-primary">{zone.payout}</span>
                  </div>
                </div>
              </div>
            ))}

            {/* ML badge */}
            <div className="rounded-xl bg-primary/8 border border-primary/15 p-4 flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-primary/20 flex items-center justify-center text-sm flex-shrink-0">🏆</div>
              <p className="text-text-secondary text-xs">
                <span className="font-semibold text-text-primary">XGBoost trained on 18 features:</span>{" "}
                zone disruption history · monsoon proximity · rolling AQI averages · earnings variance · vehicle type · cold-start indicator + 12 more
              </p>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
