import React, { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { IconChaos, IconLab, IconGlobe, IconTimer } from "./AllSectionIcons";

const features = [
  {
    id: "chaos",
    icon: <IconChaos />,
    badge: "Built for Demo",
    badgeColor: "bg-brand-primary text-white border-brand-primary/20 shadow-sm",
    title: "Chaos Engine",
    subtitle: "Simulate disruption without waiting for a real flood",
    description:
      "The platform dashboard ships with a dedicated simulation control — the Chaos Engine. Operators can inject weather signals, trigger AQI alerts, or collapse zone order volumes on demand. The full claim pipeline fires exactly as it would during a real disruption event.",
    highlights: [
      "Trigger WEATHER_ALERT, AQI_ALERT, or ORDER_DROP from UI",
      "Fires real claim generation and payout pipeline",
      "No live flood required for a complete end-to-end demo",
      "Used for load testing, fraud testing, and insurer showcases",
    ],
  },
  {
    id: "fraud",
    icon: <IconLab />,
    badge: "ML-Powered",
    badgeColor: "bg-brand-soft text-brand-primary border-brand-primary/20",
    title: "3-Layer Fraud Engine",
    subtitle: "Intercepts coordinated fraud across 6 behavioral dimensions",
    description:
      "InDel's fraud defense is economic, not geographic. GPS is spoofable — but replicating accurate earnings drops across hundreds of independent agents simultaneously is nearly impossible. Three layers work in sequence.",
    highlights: [
      "Layer 1 — IsolationForest: coordinated vectors produce short path-lengths → score >0.55 triggers hold",
      "Layer 2 — DBSCAN: worker behavior checked against zone cluster; outliers flagged",
      "Layer 3 — Postgres hard rules: delivered during disruption window? Auto-rejected",
      "Flagged claims route to manual queue with full violations[] JSON",
    ],
  },
  {
    id: "shap",
    icon: <IconGlobe />,
    badge: "Multilingual",
    badgeColor: "bg-brand-soft text-brand-primary border-brand-primary/20",
    title: "SHAP Explainability",
    subtitle: "Every premium, auditable — in the worker's language",
    description:
      "The XGBoost premium model trains on 18 features. Every output is decomposed by SHAP TreeExplainer into contributing factors, then surfaced to workers in plain language via a purpose-built 3-step translation pipeline that preserves native grammar.",
    highlights: [
      "English · Tamil · Hindi — grammatically validated static templates",
      "Icon-based visual cues for low-literacy users",
      "Maintenance Check: workers can trigger a self-service SHAP audit (max 3/day)",
      "Premium breakdowns show exact rupee contribution per risk factor",
    ],
    example: {
      title: "Your premium this week: ₹18",
      lines: [
        { label: "Flood risk in your zone", value: "+₹6", width: "66%" },
        { label: "Recent AQI pattern", value: "+₹3", width: "33%" },
        { label: "Income instability score", value: "+₹2", width: "22%" },
        { label: "Base rate", value: "₹7", width: "77%" },
      ],
    },
  },
  {
    id: "ttl",
    icon: <IconTimer />,
    badge: "Anti-Ghost Login",
    badgeColor: "bg-brand-primary text-white border-brand-primary/20",
    title: "TTL Security Gate",
    subtitle: "15-minute backward window locks eligibility at the millisecond",
    description:
      "When disruption is confirmed, the system checks every worker's lastActiveAt telemetry against a hardcoded 15-minute backward-looking window — locked to the millisecond of disruption confirmation. Dormant accounts that log in after seeing the rain are stripped from eligibility arrays automatically.",
    highlights: [
      "Hardcoded 15-min TTL — not configurable by any API parameter",
      "Evaluation timestamp locked at disruption confirmation moment",
      "Late logins auto-stripped from eligible worker arrays",
      "Works independently of the fraud ML layers",
    ],
  },
];

export default function FeatureSection() {
  const [activeId, setActiveId] = useState<string | null>(null);

  return (
    <section id="features" className="py-28 dark-section relative overflow-hidden">
      <div className="absolute inset-0 grid-overlay opacity-20 pointer-events-none" />
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[600px] h-[300px] bg-primary/10 blur-3xl rounded-full pointer-events-none" />

      <div className="relative z-10 max-w-6xl mx-auto px-6">
        {/* Header */}
        <motion.div
          className="flex justify-center mb-6"
          initial={{ opacity: 0, y: -12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
        >
          <span className="section-tag">✨ Signature Features</span>
        </motion.div>
        <motion.h2
          className="text-4xl md:text-5xl font-extrabold text-white text-center mb-4 leading-tight"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.1 }}
        >
          The systems that make
          <br />
          <span className="text-gradient-light">InDel defensible.</span>
        </motion.h2>
        <motion.p
          className="text-white/50 text-lg text-center max-w-2xl mx-auto mb-16"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.25 }}
        >
          Four proprietary systems not found in any competing parametric product.
        </motion.p>

        {/* Feature cards */}
        <div className="grid md:grid-cols-2 gap-6">
          {features.map((feat, i) => (
            <motion.div
              key={feat.id}
              initial={{ opacity: 0, y: 28 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, amount: 0.15 }}
              transition={{ duration: 0.6, delay: i * 0.1 }}
              className={`glass-dark rounded-2xl p-7 cursor-pointer transition-all duration-300 ${
                activeId === feat.id ? "ring-1 ring-primary" : "hover:ring-1 hover:ring-primary/30"
              }`}
              onClick={() => setActiveId(activeId === feat.id ? null : feat.id)}
            >
              {/* Card header */}
              <div className="flex items-start justify-between gap-4 mb-4">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl bg-primary/20 flex items-center justify-center text-xl flex-shrink-0">
                    {feat.icon}
                  </div>
                  <div>
                    <h3 className="font-bold text-white text-base">{feat.title}</h3>
                    <p className="text-white/40 text-xs mt-0.5">{feat.subtitle}</p>
                  </div>
                </div>
                <span className={`text-xs font-semibold px-2.5 py-1 rounded-full flex-shrink-0 ${feat.badgeColor}`}>
                  {feat.badge}
                </span>
              </div>

              <p className="text-white/60 text-sm leading-relaxed mb-4">{feat.description}</p>

              {/* Toggle indicator */}
              <div className="flex items-center gap-2 text-primary-light text-sm font-medium">
                <span>{activeId === feat.id ? "Hide details ↑" : "See details ↓"}</span>
              </div>

              {/* Expandable details */}
              <AnimatePresence>
                {activeId === feat.id && (
                  <motion.div
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: "auto" }}
                    exit={{ opacity: 0, height: 0 }}
                    transition={{ duration: 0.3 }}
                    className="overflow-hidden"
                  >
                    <div className="mt-5 pt-5 border-t border-white/10 space-y-2">
                      {feat.highlights.map((h, hi) => (
                        <div key={hi} className="flex items-start gap-2.5">
                          <span className="text-primary mt-1 flex-shrink-0">◆</span>
                          <span className="text-white/70 text-xs leading-relaxed">{h}</span>
                        </div>
                      ))}
                    </div>

                    {/* SHAP visual example */}
                    {feat.example && (
                      <div className="mt-5 p-4 rounded-xl bg-white/5 border border-white/10">
                        <p className="text-white/50 text-xs mb-3 font-mono">{feat.example.title}</p>
                        {feat.example.lines.map((line, li) => (
                          <div key={li} className="mb-2">
                            <div className="flex justify-between text-xs mb-1">
                              <span className="text-white/50">{line.label}</span>
                              <span className="text-primary-light font-semibold font-mono">{line.value}</span>
                            </div>
                            <div className="h-1.5 bg-white/10 rounded-full overflow-hidden">
                              <div className="h-full bg-primary rounded-full" style={{ width: line.width }} />
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </motion.div>
                )}
              </AnimatePresence>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
