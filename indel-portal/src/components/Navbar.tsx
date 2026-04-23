import React, { useState, useEffect, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";

function scrollTo(id: string) {
  const el = document.getElementById(id);
  if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
}

const navLinks = [
  { label: "Problem", id: "problem" },
  { label: "How It Works", id: "how-it-works" },
  { label: "Features", id: "features" },
  { label: "Technology", id: "tech" },
  { label: "Demo", id: "demo" },
];

export default function Navbar() {
  const [scrolled, setScrolled] = useState(false);
  const [scrollProgress, setScrollProgress] = useState(0);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [activeSection, setActiveSection] = useState("");

  useEffect(() => {
    const handleScroll = () => {
      const scrollTop = window.scrollY;
      const docHeight = document.documentElement.scrollHeight - window.innerHeight;
      setScrolled(scrollTop > 40);
      setScrollProgress(docHeight > 0 ? (scrollTop / docHeight) * 100 : 0);

      const sections = navLinks.map((l) => l.id);
      for (let i = sections.length - 1; i >= 0; i--) {
        const el = document.getElementById(sections[i]);
        if (el && el.getBoundingClientRect().top <= 140) {
          setActiveSection(sections[i]);
          break;
        }
      }
    };
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  return (
    <>
      <div className="fixed top-0 left-0 z-[60] h-1 bg-brand-primary transition-all duration-150" style={{ width: `${scrollProgress}%` }} />

      <motion.nav
        className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
          scrolled
            ? "bg-white/90 backdrop-blur-xl border-b border-gray-100 shadow-soft-xl"
            : "bg-transparent"
        }`}
        initial={{ y: -100 }}
        animate={{ y: 0 }}
      >
        <div className="max-w-7xl mx-auto px-8 py-5 flex items-center justify-between">
          <button
            onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
            className="flex items-center gap-2"
          >
            <div className="h-8 w-8 rounded-lg bg-brand-primary flex items-center justify-center">
              <span className="text-white text-base font-black italic">In</span>
            </div>
            <span className={`text-xl font-black font-['Outfit'] italic tracking-tighter ${scrolled ? "text-gray-900" : "text-gray-900"}`}>Del</span>
          </button>

          <div className="hidden md:flex items-center gap-2">
            {navLinks.map((link) => (
              <button
                key={link.id}
                onClick={() => scrollTo(link.id)}
                className={`px-5 py-2 rounded-full text-[10px] font-black uppercase tracking-[0.2em] transition-all ${
                  activeSection === link.id
                    ? "text-brand-primary"
                    : "text-gray-400 hover:text-gray-900"
                }`}
              >
                {link.label}
              </button>
            ))}
          </div>

          <div className="hidden md:flex items-center gap-4">
            <a
              href="#demo"
              onClick={(e) => { e.preventDefault(); scrollTo("demo"); }}
              className="px-6 py-2.5 rounded-full border-2 border-gray-900 text-gray-900 text-[10px] font-black uppercase tracking-widest transition-all hover:bg-brand-dark hover:border-brand-dark hover:text-white"
            >
              Watch Video
            </a>
          </div>

          <button
            className="md:hidden p-2 text-gray-900"
            onClick={() => setMobileOpen(!mobileOpen)}
          >
            {mobileOpen ? <span className="text-2xl font-light">✕</span> : <span className="text-2xl font-light">☰</span>}
          </button>
        </div>

        <AnimatePresence>
          {mobileOpen && (
            <motion.div
              className="md:hidden bg-white border-t border-gray-100 px-8 py-10 flex flex-col gap-6 shadow-2xl"
              initial={{ opacity: 0, y: -20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
            >
              {navLinks.map((link) => (
                <button
                  key={link.id}
                  onClick={() => { scrollTo(link.id); setMobileOpen(false); }}
                  className="text-left text-2xl font-black tracking-tighter text-gray-900 hover:text-brand-primary transition-colors"
                >
                  {link.label}
                </button>
              ))}
            </motion.div>
          )}
        </AnimatePresence>
      </motion.nav>
    </>
  );
}
