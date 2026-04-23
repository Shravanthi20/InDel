import React from "react";
import { motion } from "framer-motion";

const team = [
  {
    name: "Shravanthi S",
    built: "Core Policy Logic · Disruption Sync Cycle · Payout & Data Operations",
    color: "bg-brand-soft border-brand-primary/20",
  },
  {
    name: "Gayathri U",
    built: "Delivery Management · Postgres Schema · DevOps & Docker Compositions",
    color: "bg-brand-soft/60 border-brand-primary/15",
  },
  {
    name: "Rithanya K A",
    built: "Python FastAPI ML Services — XGBoost Training & Inference",
    color: "bg-brand-soft/40 border-brand-primary/10",
  },
  {
    name: "Saravana Priyaa C R",
    built: "Platform Integration · Chaos Engine · Disruption Engine",
    color: "bg-brand-soft/30 border-brand-primary/5",
  },
  {
    name: "Subiksha MV",
    built: "Insurer System · Claims Intelligence & Overall System Design",
    color: "bg-gray-50 border-gray-100",
  },
];

const visionItems = [
  { icon: "", title: "10 cities by Q3", detail: "Expand to Bangalore, Delhi, Mumbai with localized zone data" },
  { icon: "", title: "Multi-insurer API", detail: "White-label B2B deployment — any IRDAI-registered insurer can deploy InDel" },
  { icon: "", title: "iOS Worker App", detail: "Kotlin Multiplatform → native iOS support for broader reach" },
  { icon: "", title: "gRPC ML Pipeline", detail: "Replace REST with gRPC for Go ↔ Python ML at production latencies" },
];

export default function TeamSection() {
  return (
    <section id="team" className="py-28 bg-background relative overflow-hidden">
      <div className="absolute top-0 inset-x-0 h-px bg-gradient-to-r from-transparent via-primary/20 to-transparent" />

      <div className="relative z-10 max-w-6xl mx-auto px-6">
        {/* Vision */}
        <motion.div
          className="flex justify-center mb-6"
          initial={{ opacity: 0, y: -12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
        >
          <span className="section-tag bg-primary/10 text-primary border border-primary/20">
            🔭 Vision & Roadmap
          </span>
        </motion.div>

        <motion.h2
          className="text-4xl md:text-5xl font-extrabold text-center text-text-primary mb-4 leading-tight"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.1 }}
        >
          Built to scale.
          <br />
          <span className="text-gradient">Across every city. Every worker.</span>
        </motion.h2>
        <motion.p
          className="text-text-secondary text-lg text-center max-w-2xl mx-auto mb-14"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.25 }}
        >
          InDel is the default income protection layer for India's gig economy — expanding to new geographies, platforms, and worker types.
        </motion.p>

        {/* Vision items */}
        <div className="grid sm:grid-cols-2 md:grid-cols-4 gap-4 mb-24">
          {visionItems.map((v, i) => (
            <motion.div
              key={i}
              className="rounded-2xl bg-white border border-primary/10 p-5 card-hover"
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
            >
              <span className="text-2xl mb-3 block">{v.icon}</span>
              <h4 className="font-bold text-text-primary text-sm mb-1">{v.title}</h4>
              <p className="text-text-secondary text-xs leading-relaxed">{v.detail}</p>
            </motion.div>
          ))}
        </div>

        {/* Team */}
        <motion.div
          className="flex justify-center mb-8"
          initial={{ opacity: 0, y: -12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
        >
          <span className="section-tag bg-primary/10 text-primary border border-primary/20">
            👥 Team ImaginAI
          </span>
        </motion.div>

        <motion.p
          className="text-text-secondary text-center text-lg mb-12"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.1 }}
        >
          Five people. One conviction: gig workers deserve financial protection that works as hard as they do.
        </motion.p>

        <div className="grid sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          {team.map((member, i) => (
            <motion.div
              key={i}
              className={`rounded-2xl border ${member.color} p-5 text-center card-hover`}
              initial={{ opacity: 0, y: 24 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.08 }}
            >
              <div className="text-4xl mb-3">{member.emoji}</div>
              <h4 className="font-bold text-text-primary text-sm mb-2">{member.name}</h4>
              <p className="text-text-secondary text-xs leading-relaxed">{member.built}</p>
            </motion.div>
          ))}
        </div>

        {/* Hackathon badge */}
        <motion.div
          className="mt-14 flex justify-center"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.3 }}
        >
          <div className="glass rounded-full px-6 py-3 flex items-center gap-3 shadow-sm border border-primary/10">
            <span className="text-lg">🏆</span>
            <span className="text-text-primary text-sm font-semibold">Guidewire DEVTrails 2026</span>
            <span className="text-text-secondary text-xs">·</span>
            <span className="text-text-secondary text-xs">E-Commerce Delivery Persona · Phase 3: Scale & Optimize</span>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
