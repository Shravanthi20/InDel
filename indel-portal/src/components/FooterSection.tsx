import React from "react";
import { motion } from "framer-motion";

export default function FooterSection() {
  return (
    <footer className="bg-dark-900 border-t border-white/5 pt-16 pb-8">
      <div className="max-w-6xl mx-auto px-6">
        <div className="grid md:grid-cols-3 gap-10 mb-12">
          {/* Brand */}
          <div>
            <div className="font-bold text-2xl tracking-tight mb-3">
              <span className="text-primary">In</span>
              <span className="text-white">Del</span>
            </div>
            <p className="text-white/40 text-sm leading-relaxed mb-4">
              B2B parametric income insurance for India's gig delivery workers.
              Zero-touch. Zero claims. Just protection.
            </p>
            <motion.a
              href="#demo"
              onClick={(e) => { e.preventDefault(); document.getElementById("demo")?.scrollIntoView({ behavior: "smooth" }); }}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-white text-sm font-semibold hover:bg-primary-dark transition-all"
              whileHover={{ scale: 1.03 }}
            >
              ▶ Watch Demo
            </motion.a>
          </div>

          {/* Links */}
          <div className="grid grid-cols-2 gap-6">
            <div>
              <p className="text-white/30 text-xs uppercase tracking-widest font-semibold mb-4">Dashboards</p>
              <div className="space-y-2.5">
                {[
                  { label: "Platform Dashboard", link: "https://indel-platform-dashboard.onrender.com" },
                  { label: "Insurer Dashboard", link: "https://indel-urn9.onrender.com" },
                  { label: "GitHub Repo", link: "https://github.com/Shravanthi20/InDel" },
                ].map((l, i) => (
                  <a
                    key={i}
                    href={l.link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="block text-white/50 hover:text-white text-sm transition-colors"
                  >
                    {l.label} →
                  </a>
                ))}
              </div>
            </div>
            <div>
              <p className="text-white/30 text-xs uppercase tracking-widest font-semibold mb-4">Sections</p>
              <div className="space-y-2.5">
                {[
                  { label: "Problem", id: "problem" },
                  { label: "How It Works", id: "how-it-works" },
                  { label: "Features", id: "features" },
                  { label: "Technology", id: "tech" },
                  { label: "Team", id: "team" },
                ].map((l, i) => (
                  <button
                    key={i}
                    onClick={() => document.getElementById(l.id)?.scrollIntoView({ behavior: "smooth" })}
                    className="block text-white/50 hover:text-white text-sm transition-colors text-left"
                  >
                    {l.label}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Badges */}
          <div>
            <p className="text-white/30 text-xs uppercase tracking-widest font-semibold mb-4">Built for</p>
            <div className="space-y-2">
              {[
                "Guidewire DEVTrails 2026",
                "E-Commerce Delivery Persona",
                "Phase 3: Scale & Optimize",
                "Team ImaginAI",
              ].map((badge, i) => (
                <div key={i} className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-white/10 text-white/40 text-xs block w-fit mb-1">
                  {badge}
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Bottom bar */}
        <div className="border-t border-white/5 pt-6 flex flex-col md:flex-row justify-between items-center gap-4">
          <p className="text-white/20 text-xs text-center md:text-left">
            © 2026 Team ImaginAI · InDel — Insure, Deliver
          </p>
          <p className="text-white/15 text-xs text-center md:text-right max-w-lg">
            Figures are illustrative estimates for design and modelling purposes. Production deployment requires IRDAI registration and KYC/AML compliance by the deploying insurer. All systems are strictly idempotent — mass disruption events cannot produce duplicate claims or double payouts.
          </p>
        </div>
      </div>
    </footer>
  );
}
