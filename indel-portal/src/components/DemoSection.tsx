import React from "react";
import { motion } from "framer-motion";

export default function DemoSection() {
  return (
    <section id="demo" className="py-28 dark-section relative overflow-hidden">
      <div className="absolute inset-0 grid-overlay opacity-20 pointer-events-none" />

      <div className="relative z-10 max-w-5xl mx-auto px-6">
        {/* Tag */}
        <motion.div
          className="flex justify-center mb-6"
          initial={{ opacity: 0, y: -12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
        >
          <span className="section-tag">Product Demo</span>
        </motion.div>

        <motion.h2
          className="text-4xl md:text-5xl font-extrabold text-white text-center mb-4 leading-tight"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.1 }}
        >
          See it in a few minutes.
        </motion.h2>
        <motion.p
          className="text-white/50 text-lg text-center max-w-xl mx-auto mb-12"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.25 }}
        >
          Onboarding → disruption detection → fraud check → Kafka payout pipeline → UPI transfer. End-to-end, zero to payout.
        </motion.p>

        {/* Video embed */}
        <motion.div
          className="rounded-2xl overflow-hidden shadow-2xl shadow-primary/20 border border-primary/20"
          initial={{ opacity: 0, scale: 0.97 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.2 }}
          style={{ position: "relative", paddingBottom: "56.25%", height: 0 }}
        >
          <iframe
            src={import.meta.env.VITE_YOUTUBE_DEMO_URL}
            title="InDel Demo — Walkthrough"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowFullScreen
            style={{ position: "absolute", top: 0, left: 0, width: "100%", height: "100%" }}
          />
        </motion.div>

        {/* Quick access */}
        <motion.div
          className="mt-12 grid sm:grid-cols-3 gap-4"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.3 }}
        >
          {[
            { label: "Platform Dashboard", link: import.meta.env.VITE_PLATFORM_DASHBOARD_URL, desc: "Live zone telemetry & Chaos Engine" },
            { label: "Insurer Dashboard", link: import.meta.env.VITE_INSURER_DASHBOARD_URL, desc: "Loss ratio, fraud queue & Prophet forecast" },
            { label: "GitHub Repository", link: import.meta.env.VITE_GITHUB_REPO_URL, desc: "Full source code & setup guide" },
          ].map((item, i) => (
            <a
              key={i}
              href={item.link}
              target="_blank"
              rel="noopener noreferrer"
              className="glass-dark rounded-xl p-5 flex flex-col gap-2 card-hover group"
            >
              {/* <span className="text-2xl">{item.icon}</span> */}
              <span className="font-semibold text-white text-sm group-hover:text-primary-light transition-colors">{item.label} →</span>
              <span className="text-white/40 text-xs">{item.desc}</span>
            </a>
          ))}
        </motion.div>
      </div>
    </section>
  );
}
