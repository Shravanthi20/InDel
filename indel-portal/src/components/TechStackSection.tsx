import React from "react";
import { motion } from "framer-motion";
import {
  IconPhone, IconWeather, IconChaos, IconLightning, IconSearch, IconTimer, IconDatabase, IconChart, IconLab, IconCrystalBall, IconMail, IconLink, IconMoney
} from "./TechIcons";

const layers = [
  {
    label: "External Layer",
    color: "border-brand-primary/10 bg-brand-soft/40",
    labelColor: "text-brand-primary/80",
    nodes: [
      { name: "Worker Mobile Telemetry", icon: <IconPhone /> },
      { name: "Environmental Webhooks\n(IMD / AQI / OpenWeatherMap)", icon: <IconWeather /> },
      { name: "Chaos Engine\n(Telemetry Simulator)", icon: <IconChaos /> },
    ],
  },
  {
    label: "InDel Core — Go / GORM",
    color: "border-brand-primary/30 bg-brand-soft",
    labelColor: "text-brand-dark",
    nodes: [
      { name: "Gin REST Handlers & Routing", icon: <IconLightning /> },
      { name: "Disruption Engine\n(Multi-Signal Validation)", icon: <IconSearch /> },
      { name: "TTL Security Gate\n(15-min Window)", icon: <IconTimer /> },
      { name: "PostgreSQL\n(Primary Store)", icon: <IconDatabase /> },
    ],
  },
  {
    label: "ML Intelligence Layer",
    color: "border-brand-primary/20 bg-brand-soft/60",
    labelColor: "text-brand-primary",
    nodes: [
      { name: "ml-premium\nXGBoost + SHAP :9001", icon: <IconChart /> },
      { name: "ml-fraud\nIsolationForest + DBSCAN :9002", icon: <IconLab /> },
      { name: "ml-forecast\nProphet Forecasting :9003", icon: <IconCrystalBall /> },
    ],
  },
  {
    label: "Execution Layer",
    color: "border-brand-primary/10 bg-brand-soft/30",
    labelColor: "text-brand-primary/70",
    nodes: [
      { name: "Kafka Async\nPayout Engine", icon: <IconMail /> },
      { name: "Zookeeper\nCoordination", icon: <IconLink /> },
      { name: "Razorpay\nUPI Payout Rails", icon: <IconMoney /> },
    ],
  },
];

const techBadges = [
  "Go · Gin", "PostgreSQL · GORM", "React 18 · Vite · TypeScript",
  "Apache Kafka", "Apache Zookeeper", "Docker · Compose",
  "Python · FastAPI", "XGBoost · SHAP", "IsolationForest · DBSCAN",
  "Prophet", "Razorpay SDK", "Firebase FCM",
  "OpenWeatherMap", "OpenAQ · WAQI", "JWT Auth",
];

export default function TechStackSection() {
  return (
    <section id="tech" className="py-28 bg-background relative overflow-hidden">
      <div className="absolute bottom-0 left-0 right-0 h-64 bg-gradient-to-t from-primary-light/15 to-transparent pointer-events-none" />

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
            🏗️ Architecture
          </span>
        </motion.div>

        <motion.h2
          className="text-4xl md:text-5xl font-extrabold text-center text-text-primary mb-4 leading-tight"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.1 }}
        >
          One command.
          <br />
          <span className="text-gradient">Fully running.</span>
        </motion.h2>
        <motion.p
          className="text-text-secondary text-lg text-center max-w-2xl mx-auto mb-4"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.25 }}
        >
          7 containers. 3 ML microservices. Pre-seeded with workers, zones, and disruption history.
        </motion.p>

        {/* Docker compose command */}
        <motion.div
          className="max-w-2xl mx-auto mb-16"
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.35 }}
        >
          <div className="rounded-xl bg-dark-900 border border-primary/20 p-4 flex items-center gap-3">
            <span className="text-primary font-mono text-sm">$</span>
            <code className="font-mono text-sm text-green-400 flex-1 overflow-x-auto">
              COMPOSE_PARALLEL_LIMIT=1 docker compose -f docker-compose.demo.yml up --build -d
            </code>
          </div>
          <p className="text-center text-text-secondary text-xs mt-2">
            Startup order enforced: Zookeeper → Kafka → API. No manual setup. No configuration.
          </p>
        </motion.div>

        {/* Architecture layers */}
        <div className="space-y-4 mb-16">
          {layers.map((layer, i) => (
            <motion.div
              key={i}
              className={`rounded-2xl border ${layer.color} p-5`}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.55, delay: i * 0.1 }}
            >
              <p className={`text-xs uppercase tracking-widest font-bold mb-4 ${layer.labelColor}`}>
                {layer.label}
              </p>
              <div className="flex flex-wrap gap-3">
                {layer.nodes.map((node, ni) => (
                  <div
                    key={ni}
                    className="flex items-center gap-2 bg-white rounded-xl px-4 py-2.5 shadow-sm border border-white/60 text-xs text-text-primary font-medium"
                  >
                    <span className="text-base">{node.icon}</span>
                    <span className="whitespace-pre-line leading-tight">{node.name}</span>
                  </div>
                ))}
              </div>
            </motion.div>
          ))}
        </div>

        {/* Tech badges */}
        <motion.div
          className="text-center"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.1 }}
        >
          <p className="text-text-secondary text-xs uppercase tracking-widest font-semibold mb-6">Full technology roster</p>
          <div className="flex flex-wrap justify-center gap-2">
            {techBadges.map((badge, i) => (
              <motion.span
                key={i}
                className="px-3 py-1.5 rounded-full border border-primary/15 bg-white text-text-primary text-xs font-medium shadow-sm hover:bg-primary-light transition-all duration-200"
                initial={{ opacity: 0, scale: 0.9 }}
                whileInView={{ opacity: 1, scale: 1 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.04, duration: 0.3 }}
                whileHover={{ scale: 1.05 }}
              >
                {badge}
              </motion.span>
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  );
}
