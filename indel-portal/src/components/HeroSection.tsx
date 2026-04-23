import React, { useEffect, useRef, useState } from "react";
import { motion } from "framer-motion";

const stats = [
  { value: "15M+", label: "Gig Workers in India" },
  { value: "₹527", label: "Avg. Payout / Event" },
  { value: "2 min", label: "Payout Time" },
  { value: "Zero", label: "Claims to File" },
];

function AnimatedStat({ value, label, delay }: { value: string; label: string; delay: number }) {
  return (
    <motion.div
      className="flex flex-col items-center"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: delay + 0.8, duration: 0.6, ease: "easeOut" }}
    >
      <span className="text-2xl md:text-3xl font-bold text-white">{value}</span>
      <span className="text-xs md:text-sm text-white/50 mt-1 text-center">{label}</span>
    </motion.div>
  );
}

export default function HeroSection() {
  const [typedLine, setTypedLine] = useState("");
  const headline = "Verify economic loss, not location.";
  const ref = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    let i = 0;
    const tick = () => {
      setTypedLine(headline.slice(0, i + 1));
      i++;
      if (i < headline.length) ref.current = setTimeout(tick, 50);
    };
    const start = setTimeout(tick, 800);
    return () => { clearTimeout(start); if (ref.current) clearTimeout(ref.current); };
  }, []);

  return (
    <section
      id="hero"
      className="relative min-h-screen flex flex-col justify-center items-center overflow-hidden bg-white px-6 pt-20"
    >
      {/* Visual Hook: Radial Glow & Grid */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(236,72,153,0.08),transparent_70%)] pointer-events-none" />
      <div className="absolute inset-0 grid-overlay opacity-20 pointer-events-none" />

      <div className="relative z-10 max-w-5xl mx-auto text-center">
        {/* Hackathon Badge */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="inline-flex items-center gap-2 mb-10 px-4 py-2 rounded-full border border-brand-primary/20 bg-brand-soft/50 text-brand-primary text-[10px] font-black tracking-[.2em] uppercase"
        >
          <span className="w-1.5 h-1.5 rounded-full bg-brand-primary animate-pulse" />
          Guidewire DEVTrails 2026 · Project InDel
        </motion.div>

        {/* Typewriter Hook */}
        <motion.div
          className="font-mono text-sm md:text-base text-gray-400 mb-6 h-6 flex items-center justify-center gap-1"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
        >
          <span className="opacity-50">❯</span> {typedLine}
          <span className="inline-block w-1.5 h-4 bg-brand-primary animate-pulse" />
        </motion.div>

        {/* The Big Statement */}
        <motion.h1
          className="text-6xl md:text-8xl font-black leading-[0.95] tracking-tighter text-gray-900 mb-8 font-['Outfit']"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          Parametric income <br /> 
          <span className="text-brand-primary italic">that works.</span>
        </motion.h1>

        <motion.p
          className="text-lg md:text-xl text-gray-500 max-w-2xl mx-auto leading-relaxed mb-12"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.5 }}
        >
          InDel is a zero-touch insurance pipeline for the Indian gig economy. 
          When disruptions occur, we verify economic signals and trigger UPI disbursements — 
          <span className="text-gray-900 font-bold"> in under 2 minutes.</span>
        </motion.p>

        {/* CTAs following User Request */}
        <motion.div
          className="flex flex-col sm:flex-row gap-6 justify-center items-center"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6 }}
        >
          <motion.a
            href="#demo"
            className="group relative px-10 py-4 rounded-full border-2 border-gray-900 text-gray-900 font-black text-xs uppercase tracking-[0.2em] transition-all hover:bg-brand-dark hover:border-brand-dark hover:text-white overflow-hidden"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.98 }}
          >
            <span className="relative z-10 font-black">Watch Live Reveal</span>
          </motion.a>
          
          <motion.a
            href="#ecosystem"
            className="text-gray-400 hover:text-gray-900 text-xs font-black uppercase tracking-[0.2em] transition-colors flex items-center gap-2"
          >
            Explore Ecosystem <span className="text-lg">→</span>
          </motion.a>
        </motion.div>
      </div>

      {/* Hero Stats */}
      <motion.div 
        className="mt-20 w-full max-w-4xl grid grid-cols-2 md:grid-cols-4 gap-8 px-6"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.8 }}
      >
        {stats.map((s, i) => (
          <div key={i} className="text-center group py-4">
            <div className="text-3xl font-black text-gray-900 font-['Outfit'] group-hover:text-brand-primary transition-colors">{s.value}</div>
            <div className="text-[10px] font-black uppercase tracking-[0.2em] text-gray-400 mt-1">{s.label}</div>
          </div>
        ))}
      </motion.div>

      {/* Scroll indicator */}
      <motion.div
        className="absolute bottom-6 left-1/2 -translate-x-1/2 flex flex-col items-center gap-1 text-gray-300 text-xs"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1.6, duration: 0.5 }}
      >
        <div className="w-5 h-8 border border-gray-200 rounded-full flex justify-center pt-1.5">
          <motion.div
            className="w-1 h-1.5 bg-brand-primary/40 rounded-full"
            animate={{ y: [0, 10, 0] }}
            transition={{ duration: 1.5, repeat: Infinity, ease: "easeInOut" }}
          />
        </div>
        scroll
      </motion.div>
    </section>
  );
}
