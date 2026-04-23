import React from "react";
import { motion } from "framer-motion";
import { IconPhone, IconMap, IconChart, IconPuzzle } from "./AllSectionIcons";

const products = [
  {
    icon: <IconPhone />,
    title: "Worker App",
    subtitle: "For delivery partners",
    description: "Native Android app. Everything on one screen — no hunting through menus.",
    features: [
      "Coverage status & policy state (Active / Paused / Rewarded)",
      "AI-computed weekly premium with SHAP breakdown",
      "Earnings vs protected baseline chart",
      "Active disruption alerts in real time",
      "Claim history + continuity reward progress",
      "Maintenance Check — self-service SHAP audit (3/day)",
      "Razorpay UPI payment — one tap",
    ],
    link: "https://github.com/Shravanthi20/InDel#worker-app",
    linkLabel: "View on GitHub →",
    badge: "Android · Kotlin",
    badgeColor: "bg-brand-soft text-brand-primary border-brand-primary/20",
    highlight: "SHAP · Razorpay · Firebase FCM",
    highlightColor: "border-brand-primary/10",
  },
  {
    icon: <IconMap />,
    title: "Platform Dashboard",
    subtitle: "For operators & platform admins",
    description: "Real-time zone telemetry and the Chaos Engine — simulate disruptions without waiting for actual events.",
    features: [
      "Live zone telemetry & order velocity heatmap",
      "Worker GPS distribution by zone",
      "Chaos Engine — inject WEATHER_ALERT, AQI drops, zone closures",
      "Disruption event log with multi-signal confirmation status",
      "Worker eligibility scan results per event",
      "Claim pipeline trigger & monitoring",
    ],
    link: "https://indel-platform-dashboard.onrender.com",
    linkLabel: "Open Dashboard →",
    badge: "React · Vite · TypeScript",
    badgeColor: "bg-brand-soft text-brand-primary border-brand-primary/20",
    highlight: "Chaos Engine · Live Telemetry",
    highlightColor: "border-brand-primary/10",
  },
  {
    icon: <IconChart />,
    title: "Insurer Dashboard",
    subtitle: "For insurance providers",
    description: "Every number an actuary needs — live. Reserve planning, fraud queue, and 7-day Prophet disruption forecast.",
    features: [
      "Premium pool health & loss ratio by zone and city",
      "Live fraud queue with full violations[] JSON inline",
      "Prophet Reserve Analytics — 7-day claim volume forecast",
      "OpenWeatherMap-seeded disruption probability per zone",
      "Claim approval, hold, and manual review workflow",
      "Insurer review notes & escalation tracking",
    ],
    link: "https://indel-urn9.onrender.com",
    linkLabel: "Open Dashboard →",
    badge: "Tremor · React · Vite",
    badgeColor: "bg-brand-primary text-white border-brand-primary/20 shadow-sm",
    highlight: "Prophet · IsolationForest · DBSCAN",
    highlightColor: "border-brand-primary/20 shadow-lg shadow-brand-primary/5",
  },
];

const containerVariants = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.15 } },
};

const cardVariants = {
  hidden: { opacity: 0, y: 32 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.65, ease: "easeOut" } },
};

export default function ProductEcosystemSection() {
  return (
    <section id="ecosystem" className="py-28 bg-background relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-t from-primary-light/20 to-transparent pointer-events-none" />

      <div className="relative z-10 max-w-7xl mx-auto px-6">
        {/* Tag */}
        <motion.div
          className="flex justify-center mb-6"
          initial={{ opacity: 0, y: -12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
        >
          <span className="section-tag bg-primary/10 text-primary border border-primary/20">
            <IconPuzzle /> Product Ecosystem
          </span>
        </motion.div>

        <motion.h2
          className="text-4xl md:text-5xl font-extrabold text-center text-text-primary mb-4 leading-tight"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.1 }}
        >
          Three surfaces.
          <br />
          <span className="text-gradient">One unified data layer.</span>
        </motion.h2>
        <motion.p
          className="text-text-secondary text-lg text-center max-w-2xl mx-auto mb-16"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.25 }}
        >
          Worker, operator, and insurer each see exactly what they need — powered by the same InDel data infrastructure underneath.
        </motion.p>

        {/* Product cards */}
        <motion.div
          className="grid md:grid-cols-3 gap-6"
          variants={containerVariants}
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.1 }}
        >
          {products.map((p, i) => (
            <motion.div
              key={i}
              variants={cardVariants}
              className={`bg-white rounded-2xl border ${p.highlightColor} shadow-sm flex flex-col card-hover overflow-hidden`}
            >
              {/* Card header */}
              <div className="p-6 pb-4 border-b border-slate-50">
                <div className="flex items-center justify-between mb-4">
                  <span className="text-3xl">{p.icon}</span>
                  <span className={`text-xs font-semibold px-2.5 py-1 rounded-full ${p.badgeColor}`}>{p.badge}</span>
                </div>
                <h3 className="font-extrabold text-text-primary text-xl mb-1">{p.title}</h3>
                <p className="text-text-secondary text-xs font-medium uppercase tracking-wider">{p.subtitle}</p>
                <p className="text-text-secondary text-sm mt-3 leading-relaxed">{p.description}</p>
              </div>

              {/* Features list */}
              <div className="p-6 flex-1">
                <p className="text-xs uppercase tracking-widest text-text-secondary font-semibold mb-3">What's inside</p>
                <ul className="space-y-2">
                  {p.features.map((feat, fi) => (
                    <li key={fi} className="flex items-start gap-2 text-sm text-text-secondary">
                      <span className="text-primary mt-0.5 flex-shrink-0 text-xs">◆</span>
                      {feat}
                    </li>
                  ))}
                </ul>
              </div>

              {/* Highlight + CTA */}
              <div className="p-6 pt-0 flex flex-col gap-3">
                <div className="rounded-lg bg-primary-light/60 px-3 py-2 text-xs text-primary font-semibold">
                  Key tech: {p.highlight}
                </div>
                <a
                  href={p.link}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="w-full text-center px-5 py-3 rounded-xl bg-primary text-white text-sm font-semibold shadow-md shadow-primary/20 hover:bg-primary-dark transition-all"
                >
                  {p.linkLabel}
                </a>
              </div>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
}
