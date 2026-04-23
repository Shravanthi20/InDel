import React from "react";
import { motion } from "framer-motion";
import { IconSpark, IconStatWorker, IconStatChart, IconStatBan } from "./AllSectionIcons";

const problems = [
  {
    stat: "15M+",
    label: "Gig delivery workers in India",
    detail: "Every rupee earned depends on one thing: completing orders. One event wipes it out.",
    icon: <IconStatWorker />,
  },
  {
    stat: "20–30%",
    label: "Monthly income lost per disruption",
    detail: "Floods. Hazardous AQI. Curfews. Zone closures. When any of these hit — income collapses.",
    icon: <IconStatChart />,
  },
  {
    stat: "₹0",
    label: "Insurance products that cover this",
    detail: "Traditional insurance covers accidents and vehicles. Nothing covers lost earnings from disruption.",
    icon: <IconStatBan />,
  },
];

const failurePoints = [
  "Insurers request data from delivery platforms",
  "Platforms deny access — zero commercial incentive to share",
  "No data → weak verification → rampant fraud",
  "No product survives at scale",
];

const containerVariants = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.15 } },
};

const cardVariants = {
  hidden: { opacity: 0, y: 32 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.65, ease: "easeOut" } },
};

export default function ProblemSection() {
  return (
    <section id="problem" className="py-28 dark-section relative overflow-hidden">
      {/* Background accent */}
      <div className="absolute inset-0 grid-overlay opacity-25 pointer-events-none" />
      <div className="absolute -top-32 right-0 w-96 h-96 rounded-full bg-error/5 blur-3xl pointer-events-none" />

      <div className="relative z-10 max-w-6xl mx-auto px-6">
        {/* Section tag */}
        <motion.div
          className="flex justify-center mb-6"
          initial={{ opacity: 0, y: -12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
        >
          <span className="section-tag bg-error/15 text-red-300 border border-error/20">
            <IconSpark /> The Problem Nobody Has Solved
          </span>
        </motion.div>

        {/* Heading */}
        <motion.h2
          className="text-4xl md:text-5xl font-extrabold text-white text-center mb-4 leading-tight"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.1 }}
        >
          When the rain falls,
          <br />
          <span className="text-gradient-light">the orders stop.</span>
        </motion.h2>

        <motion.p
          className="text-white/50 text-center text-lg max-w-2xl mx-auto mb-16"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.25 }}
        >
          India's gig economy runs on fragile income. Workers bear the full cost of disruptions they can't predict and didn't cause.
        </motion.p>

        {/* Problem cards */}
        <motion.div
          className="grid md:grid-cols-3 gap-6 mb-20"
          variants={containerVariants}
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, amount: 0.2 }}
        >
          {problems.map((p, i) => (
            <motion.div
              key={i}
              variants={cardVariants}
              className="glass-dark rounded-2xl p-8 card-hover flex flex-col"
            >
              <span className="text-4xl mb-4">{p.icon}</span>
              <span className="text-4xl font-extrabold text-white mb-1">{p.stat}</span>
              <span className="text-primary-light font-semibold text-sm mb-3">{p.label}</span>
              <p className="text-white/50 text-sm leading-relaxed">{p.detail}</p>
            </motion.div>
          ))}
        </motion.div>

        {/* Why existing solutions fail */}
        <motion.div
          className="max-w-3xl mx-auto"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.65, delay: 0.2 }}
        >
          <p className="text-white/40 text-center text-xs uppercase tracking-widest font-semibold mb-8">
            Why existing parametric attempts have failed
          </p>
          <div className="relative pl-6">
            {/* Vertical line */}
            <div className="absolute left-2.5 top-2 bottom-2 w-px bg-white/10" />
            {failurePoints.map((point, i) => (
              <div key={i} className="relative mb-6 last:mb-0 flex items-start gap-4">
                <div className="absolute -left-6 flex-shrink-0 w-5 h-5 rounded-full bg-error/20 border border-error/40 flex items-center justify-center text-red-400 text-xs font-bold mt-0.5">
                  {i + 1}
                </div>
                <p className="text-white/60 text-sm leading-relaxed pl-8">{point}</p>
              </div>
            ))}
          </div>

          <motion.div
            className="mt-10 p-5 rounded-xl border border-error/20 bg-error/6"
            initial={{ opacity: 0, scale: 0.97 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            <p className="text-red-300 text-sm text-center font-medium">
              The result: unverifiable claims, rampant fraud, and no insurance product that works at scale.
            </p>
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
}
