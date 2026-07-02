import React from 'react';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import { motion } from 'framer-motion';

function HomepageHeader() {
  return (
    <header className="hero-gradient-bg" style={{ minHeight: '80vh', display: 'flex', alignItems: 'center', justifyContent: 'center', textAlign: 'center', padding: '4rem 1rem' }}>
      <div className="container">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, ease: "easeOut" }}
        >
          <img src="/img/logo.svg" alt="Aeroflare Logo" style={{ width: '100px', marginBottom: '2rem' }} />
          <h1 style={{ fontSize: '4rem', fontWeight: 800, marginBottom: '1.5rem', letterSpacing: '-0.02em', lineHeight: 1.1, color: '#fff' }}>
            High-Performance OCI-Backed <br />
            <span className="gradient-text">Nix Binary Cache</span> in Go
          </h1>
          <p style={{ fontSize: '1.25rem', color: '#a3aed2', maxWidth: '700px', margin: '0 auto 2.5rem' }}>
            Bridges the Nix ecosystem and standard container registries (like GHCR) to act as a stateless, zero-infrastructure binary substituter.
          </p>
          
          <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', flexWrap: 'wrap', marginBottom: '3rem' }}>
            <Link className="hero-button" to="/docs/tutorials/quick-start">
              Get Started
            </Link>
            <a className="hero-button-secondary" href="https://github.com/ItzEmoji/aeroflare" target="_blank" rel="noopener noreferrer">
              View on GitHub
            </a>
          </div>

          <motion.div 
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.6, delay: 0.3 }}
            style={{ maxWidth: '600px', margin: '0 auto' }}
          >
            <div className="terminal-block">
              <span className="terminal-prompt">$</span> <span>nix run github:ItzEmoji/aeroflare -- init</span>
            </div>
          </motion.div>
        </motion.div>
      </div>
    </header>
  );
}

function HomepageFeatures() {
  const features = [
    {
      title: 'Stateless Proxying',
      description: 'Retains zero local binary state. Streams .nar blobs directly from OCI.',
    },
    {
      title: 'O(1) Manifest Lookups',
      description: 'Tags artifacts directly with the 32-character Nix store path hash, enabling instantaneous lookups.',
    },
    {
      title: 'Dual-Backend Support',
      description: 'Use OCI registries for heavy NAR blobs and Cloudflare R2 for fast metadata (.narinfo).',
    },
    {
      title: 'Interactive Provisioning',
      description: 'Built-in setup wizard for GitHub, GitLab, and Cloudflare R2 bucket configuration.',
    },
  ];

  return (
    <section style={{ padding: '6rem 0', backgroundColor: '#0b0f19' }}>
      <div className="container">
        <h2 style={{ textAlign: 'center', fontSize: '2.5rem', fontWeight: 700, color: '#fff', marginBottom: '4rem' }}>
          Core Capabilities
        </h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '2rem' }}>
          {features.map((feature, i) => (
            <motion.div
              key={i}
              className="glass-card"
              style={{ padding: '2rem' }}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
            >
              <h3 style={{ fontSize: '1.4rem', color: '#00d2ff', fontWeight: 600, marginBottom: '1rem' }}>{feature.title}</h3>
              <p style={{ color: '#a3aed2', lineHeight: 1.6, margin: 0 }}>{feature.description}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

export default function Home(): JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="The OCI-Nix-Binary-Cache written in Go">
      <HomepageHeader />
      <HomepageFeatures />
    </Layout>
  );
}
