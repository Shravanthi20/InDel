import React, { useRef } from "react";
import { motion, useInView } from "framer-motion";

const steps = [
  {
    id: 1,
    icon: "📡",
    title: "Environmental Signal",
    detail: "OpenWeatherMap / AQI / zone closure fires a structured webhook event into the InDel disruption pipeline.",
    tag: "External Trigger",
    tagColor: "bg-brand-soft text-brand-primary border-brand-primary/20",
  },
  {
    id: 2,
    icon: "📊",
    title: "Order Velocity Collapse",
    detail: "Zone order volume drops >30% versus the 4-week sliding baseline — confirmed by InDel's own telemetry, not platform data.",
    tag: "Internal Signal",
    tagColor: "bg-brand-soft text-brand-primary border-brand-primary/20",
  },
  {
    id: 3,
    icon: "🔒",
    title: "Multi-Signal Lock",
    detail: "Both signals must align within a time-lag window. Rain with no delivery impact? Nothing fires. Order drop under clear skies? Nothing fires.",
    tag: "Dual-Lock Validation",
    tagColor: "bg-brand-primary text-white border-brand-primary/20 shadow-md",
  },
  {
    id: 4,
    icon: "🕒",
    title: "TTL Eligibility Scan",
    detail: "Each worker's lastActiveAt timestamp is checked against a hardcoded 15-minute backward window. Ghost logins stripped automatically.",
    tag: "Anti-Fraud Gate",
    tagColor: "bg-brand-soft text-brand-primary border-brand-primary/20",
  },
  {
    id: 5,
    icon: "🧠",
    title: "3-Layer Fraud Check",
    detail: "IsolationForest anomaly scoring → DBSCAN spatial cluster check → Postgres hard rules. Low-risk: instant. High-risk: manual queue with full violations JSON.",
    tag: "ML Fraud Engine",
    tagColor: "bg-brand-soft text-brand-primary border-brand-primary/20",
  },
  {
    id: 6,
    icon: "💸",
    title: "Kafka → Razorpay → UPI",
    detail: "Approved claim queued as a Kafka event. 5× exponential retry. Idempotency key prevents duplicates. Worker receives money — same day.",
    tag: "Instant Payout",
    tagColor: "bg-brand-primary text-white border-brand-primary/20 shadow-md",
  },
];

function FlowStep({ step, index, total }: { step: typeof steps[0]; index: number; total: number }) {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, { once: true, margin: "-80px" });

  return (
    <div ref={ref} className="relative flex gap-6">
      {/* Connector line */}
      {index < total - 1 && (
        <div className="absolute left-[22px] top-12 bottom-0 w-0.5 bg-primary/20 z-0">
          <motion.div
            className="w-full bg-primary origin-top"
            style={{ height: "100%" }}
            initial={{ scaleY: 0 }}
            animate={inView ? { scaleY: 1 } : { scaleY: 0 }}
            transition={{ duration: 0.8, delay: 0.4, ease: "easeOut" }}
          />
        </div>
      )}

      {/* Step circle */}
      <div className="relative z-10 flex-shrink-0">
        <motion.div
          className={`w-11 h-11 rounded-full flex items-center justify-center text-lg border-2 transition-all duration-500 ${
            inView ? "bg-primary border-primary shadow-lg shadow-primary/30" : "bg-white border-primary/20"
          }`}
          initial={{ scale: 0.7, opacity: 0 }}
          animate={inView ? { scale: 1, opacity: 1 } : { scale: 0.7, opacity: 0 }}
          transition={{ duration: 0.45, delay: index * 0.1 }}
        >
          <span>{step.icon}</span>
        </motion.div>
      </div>

      {/* Content card */}
      <motion.div
        className="flex-1 pb-10"
        initial={{ opacity: 0, x: 20 }}
        animate={inView ? { opacity: 1, x: 0 } : { opacity: 0, x: 20 }}
        transition={{ duration: 0.55, delay: index * 0.1 + 0.15 }}
      >
        <div className="bg-white rounded-2xl border border-primary/10 p-5 card-hover">
          <div className="flex items-start justify-between gap-3 mb-2">
            <h3 className="font-bold text-text-primary text-base">{step.title}</h3>
            <span className={`text-xs font-semibold px-2.5 py-1 rounded-full flex-shrink-0 ${step.tagColor}`}>
              {step.tag}
            </span>
          </div>
          <p className="text-text-secondary text-sm leading-relaxed">{step.detail}</p>
        </div>
      </motion.div>
    </div>
  );
}

export default function HowItWorksSection() {
  return (
    <section id="how-it-works" className="py-28 bg-background relative overflow-hidden">
      <div className="absolute top-0 right-0 w-1/2 h-full bg-gradient-to-l from-primary-light/20 to-transparent pointer-events-none" />

      <div className="relative z-10 max-w-6xl mx-auto px-6">
        <div className="grid lg:grid-cols-2 gap-16 items-start">
          {/* Left: Intro */}
          <div className="lg:sticky lg:top-28">
            <motion.div
              initial={{ opacity: 0, y: -12 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
              className="mb-6"
            >
              <span className="section-tag bg-primary/10 text-primary border border-primary/20">
                ⚙️ Zero-Touch Claim Flow
              </span>
            </motion.div>

            <motion.h2
              className="text-4xl md:text-5xl font-extrabold text-text-primary mb-6 leading-tight"
              initial={{ opacity: 0, y: 24 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.65, delay: 0.1 }}
            >
              Workers do
              <br />
              <span className="text-gradient">absolutely nothing.</span>
            </motion.h2>

            <motion.p
              className="text-text-secondary text-lg leading-relaxed mb-8"
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.25 }}
            >
              The most important UX decision in InDel: workers never file a claim. Six automated
              stages execute from signal receipt to UPI transfer — without a single user action.
            </motion.p>

            {/* Live example callout */}
            <motion.div
              className="rounded-2xl bg-dark-900 p-6 font-mono text-sm border border-primary/20"
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: 0.3 }}
            >
              <p className="text-white/40 text-xs uppercase tracking-widest mb-4">
                Real event — Tambaram Flood, Chennai
              </p>
              {[
                ["Disruption window", "11:40 AM → 5:30 PM"],
                ["Worker baseline", "₹120 / hour"],
                ["Expected earnings", "₹700 (5.83 hrs)"],
                ["Actual earnings", "₹80 (2 partial orders)"],
                ["Income loss", "₹620"],
                ["Coverage ratio", "85%"],
                ["", ""],
                ["Payout dispatched", "₹527 → UPI"],
              ].map(([k, v], i) =>
                k === "" ? (
                  <div key={i} className="my-3 border-t border-white/10" />
                ) : (
                  <div key={i} className="flex justify-between gap-4 mb-1.5">
                    <span className="text-white/40">{k}</span>
                    <span className={`font-semibold ${k === "Payout dispatched" ? "text-primary-light" : "text-white/80"}`}>{v}</span>
                  </div>
                )
              )}
              <div className="mt-4 pt-3 border-t border-white/10 text-white/30 text-xs italic">
                The worker received a notification. They never opened a form.
              </div>
            </motion.div>
          </div>

          {/* Right: Animated flow steps */}
          <div className="mt-4">
            {steps.map((step, i) => (
              <FlowStep key={step.id} step={step} index={i} total={steps.length} />
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
