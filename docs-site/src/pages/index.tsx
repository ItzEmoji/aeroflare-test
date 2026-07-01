import React from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import { motion } from 'framer-motion';
import { Terminal, Zap, Shield, Globe, Box, Settings } from 'lucide-react';

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className="hero-gradient-bg" style={{ padding: '8rem 0', textAlign: 'center' }}>
      <div className="container">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
        >
          <h1 className="hero__title" style={{ fontSize: '4rem', fontWeight: 700, marginBottom: '1rem', letterSpacing: '-0.02em' }}>
            The OCI-Native <br /> <span className="gradient-text">Nix Binary Cache</span>
          </h1>
          <p className="hero__subtitle" style={{ fontSize: '1.4rem', color: '#a3aed2', maxWidth: '800px', margin: '0 auto 2.5rem', lineHeight: 1.5 }}>
            {siteConfig.tagline}. Push, pull, and proxy your Nix artifacts using standard container registries like GHCR or Cloudflare R2—no database required.
          </p>
          
          <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', marginBottom: '4rem' }}>
            <Link
              className="hero-button"
              to="/docs/tutorials/quick-start">
              Get Started in 5m
            </Link>
            <Link
              className="hero-button-secondary"
              to="/docs/explanation/architecture">
              How it works
            </Link>
          </div>

          <motion.div 
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.6, delay: 0.2 }}
            style={{ maxWidth: '700px', margin: '0 auto' }}
          >
            <div className="terminal-block">
              <div><span className="terminal-prompt">$</span> <span style={{ color: '#fff' }}>nix run github:ItzEmoji/aeroflare -- init</span></div>
              <div style={{ color: '#8b949e', marginTop: '0.5rem' }}># Provision Cloudflare R2 or OCI Registry securely...</div>
              <div style={{ marginTop: '1rem' }}><span className="terminal-prompt">$</span> <span style={{ color: '#fff' }}>nix run github:ItzEmoji/aeroflare -- run -- nix build .#default</span></div>
              <div style={{ color: '#8b949e', marginTop: '0.5rem' }}># Builds locally and pushes OCI layers instantly</div>
            </div>
          </motion.div>
        </motion.div>
      </div>
    </header>
  );
}

const FeatureList = [
  {
    title: 'Stateless Architecture',
    icon: <Globe size={32} color="#00d2ff" />,
    description: (
      <>
        Aeroflare operates entirely without a local database. It maps Nix <code>.narinfo</code> metadata directly onto OCI Manifest Annotations.
      </>
    ),
  },
  {
    title: 'Blazing Fast Execution',
    icon: <Zap size={32} color="#00d2ff" />,
    description: (
      <>
        Written in Go and designed for concurrency. The <code>run</code> wrapper intercepts Nix daemon requests and streams layers in parallel.
      </>
    ),
  },
  {
    title: 'Native Cloudflare R2',
    icon: <Box size={32} color="#00d2ff" />,
    description: (
      <>
        Prefer object storage? Use the interactive wizard to automatically provision a Cloudflare R2 bucket and edge Worker in seconds.
      </>
    ),
  },
  {
    title: 'Zero-Config GHCR',
    icon: <Settings size={32} color="#00d2ff" />,
    description: (
      <>
        Seamlessly push your Nix cache to GitHub Container Registry using your standard Personal Access Token and GitHub Actions.
      </>
    ),
  },
  {
    title: 'O(1) Manifest Lookups',
    icon: <Terminal size={32} color="#00d2ff" />,
    description: (
      <>
        By enforcing a strict Nix Hash tagging schema, Aeroflare fetches metadata instantly without needing search indexes.
      </>
    ),
  },
  {
    title: 'Secure Keychain',
    icon: <Shield size={32} color="#00d2ff" />,
    description: (
      <>
        Tokens are never stored in plain text. Aeroflare integrates natively with your OS keychain to keep your Cloudflare and GitHub secrets safe.
      </>
    ),
  },
];

function Feature({title, icon, description}: {title: string, icon: React.ReactNode, description: JSX.Element}) {
  return (
    <div className={clsx('col col--4')} style={{ marginBottom: '2rem' }}>
      <div className="glass-card" style={{ padding: '2rem', height: '100%', borderRadius: '16px' }}>
        <div style={{ marginBottom: '1.5rem' }}>{icon}</div>
        <h3 style={{ fontSize: '1.5rem', marginBottom: '1rem', color: '#fff' }}>{title}</h3>
        <p style={{ color: '#a3aed2', lineHeight: 1.6, margin: 0 }}>{description}</p>
      </div>
    </div>
  );
}

export default function Home(): JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="High-performance OCI-backed Nix binary cache proxy">
      <HomepageHeader />
      <main style={{ padding: '6rem 0', backgroundColor: '#0b0f19' }}>
        <div className="container">
          <div className="row">
            {FeatureList.map((props, idx) => (
              <Feature key={idx} {...props} />
            ))}
          </div>
        </div>
      </main>
    </Layout>
  );
}
