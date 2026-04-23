import React from "react";
import { motion } from "framer-motion";

export default function InsightSection() {
  return (
    <section id="insight" className="py-28 bg-background relative overflow-hidden">
      {/* Subtle background gradient */}
      <div className="absolute inset-0 bg-gradient-to-br from-primary-light/30 via-transparent to-transparent pointer-events-none" />

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
            💡 The Core Insight
          </span>
        </motion.div>

        <motion.h2
          className="text-4xl md:text-5xl font-extrabold text-center text-text-primary mb-4 leading-tight"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.1 }}
        >
          Ask a different question.
        </motion.h2>
        <motion.p
          className="text-text-secondary text-lg text-center max-w-2xl mx-auto mb-16"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.25 }}
        >
          Every old parametric system asks: <em>"Was the worker inside the disrupted zone?"</em><br />
          GPS is trivially spoofed. That question is the wrong one.
        </motion.p>

        {/* Comparison blocks */}
        <div className="grid md:grid-cols-2 gap-6 mb-16">
          {/* Old way */}
          <motion.div
            className="rounded-2xl border border-error/20 bg-red-50 p-8"
            initial={{ opacity: 0, x: -24 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6, delay: 0.1 }}
          >
            <div className="flex items-center gap-3 mb-6">
              <span className="text-2xl">❌</span>
              <span className="font-bold text-red-700 text-sm uppercase tracking-wider">Old Way</span>
            </div>
            <div className="space-y-3 font-mono text-sm text-red-800/70">
              {[
                "Insurer → requests data from Amazon / Flipkart",
                "→ access denied",
                "→ weak verification",
                "→ fraud at scale",
                "→ no product launches",
              ].map((line, i) => (
                <div key={i} className={`${i > 0 ? "pl-4" : ""}`}>{line}</div>
              ))}
            </div>
          </motion.div>

          {/* InDel way */}
          <motion.div
            className="rounded-2xl border border-success/20 bg-green-50 p-8"
            initial={{ opacity: 0, x: 24 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6, delay: 0.2 }}
          >
            <div className="flex items-center gap-3 mb-6">
              <span className="text-2xl">✅</span>
              <span className="font-bold text-green-700 text-sm uppercase tracking-wider">InDel Way</span>
            </div>
            <div className="space-y-3 font-mono text-sm text-green-800/70">
              {[
                "Insurer deploys InDel platform",
                "→ integrates with delivery API",
                "→ first-party data layer owned by InDel",
                "→ verified economic disruption",
                "→ automated payout → zero manual claims",
              ].map((line, i) => (
                <div key={i} className={`${i > 0 ? "pl-4" : ""}`}>{line}</div>
              ))}
            </div>
          </motion.div>
        </div>

        {/* The real insight */}
        <motion.div
          className="max-w-4xl mx-auto rounded-2xl bg-primary p-10 text-center shadow-xl shadow-primary/20"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.2 }}
        >
          <p className="text-white/60 text-sm uppercase tracking-widest font-semibold mb-4">InDel's verification model</p>
          <p className="text-2xl md:text-3xl font-bold text-white leading-snug">
            "Did this worker's <span className="text-primary-light">economic reality collapse</span>?"
          </p>
          <p className="mt-6 text-white/60 max-w-2xl mx-auto leading-relaxed">
            A heatwave with no delivery impact? Triggers nothing. An order slump under clear skies? Triggers nothing.
            InDel verifies <strong className="text-white">economic reality</strong> — not atmospheric conditions.
          </p>
        </motion.div>

        {/* 5 signal types */}
        <motion.div
          className="mt-16"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.15 }}
        >
          <p className="text-center text-text-secondary text-sm uppercase tracking-widest font-semibold mb-8">
            Five signals run simultaneously. One signal is never enough.
          </p>
          <div className="grid sm:grid-cols-2 lg:grid-cols-5 gap-4">
            {[
              { code: "WEATHER_ALERT", source: "OpenWeatherMap", icon: "🌧️" },
              { code: "AQI_ALERT", source: "OpenAQ / WAQI", icon: "🏭" },
              { code: "ORDER_DROP", source: "InDel Telemetry", icon: "📦" },
              { code: "ZONE_CLOSURE", source: "Traffic / Govt APIs", icon: "🚧" },
              { code: "ACTIVITY_ANOMALY", source: "InDel Platform", icon: "📡" },
            ].map((sig, i) => (
              <div
                key={i}
                className="rounded-xl border border-primary/15 bg-white p-4 flex flex-col items-center text-center card-hover"
              >
                <span className="text-2xl mb-2">{sig.icon}</span>
                <span className="text-xs font-mono font-semibold text-primary mb-1">{sig.code}</span>
                <span className="text-xs text-text-secondary">{sig.source}</span>
              </div>
            ))}
          </div>
          <p className="text-center text-text-secondary text-sm mt-4 italic">
            A disruption is confirmed <strong>only</strong> when an external environmental signal + internal order volume collapse align simultaneously within a time-lag window.
          </p>
        </motion.div>
      </div>
    </section>
  );
}
