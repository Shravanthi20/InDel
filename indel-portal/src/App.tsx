import React from "react";
import { motion } from "framer-motion";
import Navbar from "./components/Navbar";
import HeroSection from "./components/HeroSection";
import ProblemSection from "./components/ProblemSection";
import InsightSection from "./components/InsightSection";
import HowItWorksSection from "./components/HowItWorksSection";
import FeatureSection from "./components/FeatureSection";
import PricingIntelSection from "./components/PricingIntelSection";
import PilotNumbersSection from "./components/PilotNumbersSection";
import ProductEcosystemSection from "./components/ProductEcosystemSection";
import TechStackSection from "./components/TechStackSection";
import TeamSection from "./components/TeamSection";
import DemoSection from "./components/DemoSection";
import FooterSection from "./components/FooterSection";
import "./styles/globals.css";

export default function App() {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
    >
      <Navbar />
      <main>
        {/* Act 1 — Hook */}
        <HeroSection />

        {/* Act 2 — Problem */}
        <ProblemSection />

        {/* Act 3 — Insight (why InDel is different) */}
        <InsightSection />

        {/* Act 4 — How It Works (zero-touch pipeline) */}
        <HowItWorksSection />

        {/* Act 5A — Signature Features */}
        <FeatureSection />

        {/* Act 5B — Dynamic Pricing Intelligence */}
        <PricingIntelSection />

        {/* Act 6A — Pilot Numbers */}
        <PilotNumbersSection />

        {/* Act 6B — Product Ecosystem (3 dashboards) */}
        <ProductEcosystemSection />

        {/* Act 6C — Architecture */}
        <TechStackSection />

        {/* Act 7A — Demo video */}
        <DemoSection />

        {/* Act 7B — Team + Vision */}
        <TeamSection />
      </main>
      <FooterSection />
    </motion.div>
  );
}
