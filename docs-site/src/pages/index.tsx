import React from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import { motion } from 'framer-motion';
import { Terminal, Zap, Shield, Globe, Box, Settings, HardDrive, Cpu, GitMerge } from 'lucide-react';

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
          <img src="/img/logo.svg" alt="Aeroflare Logo" style={{ width: '120px', marginBottom: '2rem' }} />
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
            style={{ maxWidth: '750px', margin: '0 auto' }}
          >
            <div className="terminal-block">
              <div><span className="terminal-prompt">$</span> <span style={{ color: '#fff' }}>nix run github:ItzEmoji/aeroflare -- init</span></div>
              <div style={{ color: '#8b949e', marginTop: '0.5rem' }}># Provision Cloudflare R2 or OCI Registry securely with interactive auth</div>
              <div style={{ marginTop: '1rem' }}><span className="terminal-prompt">$</span> <span style={{ color: '#fff' }}>nix run github:ItzEmoji/aeroflare -- run -- nix build .#default</span></div>
              <div style={{ color: '#8b949e', marginTop: '0.5rem' }}># Builds locally, intercepts Nix daemon, and streams OCI layers instantly</div>
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
        Aeroflare operates entirely without a local database. It maps Nix <code>.narinfo</code> metadata directly onto OCI Manifest Annotations (<code>vnd.aeroflare.nar.*</code>), relying on the remote registry as the ultimate source of truth.
      </>
    ),
  },
  {
    title: 'Blazing Fast Execution',
    icon: <Zap size={32} color="#00d2ff" />,
    description: (
      <>
        Written in Go and designed for concurrency. The <code>run</code> wrapper spins up an ephemeral proxy, intercepts Nix daemon requests, and streams layers in parallel with zero disk IO overhead.
      </>
    ),
  },
  {
    title: 'Native Cloudflare R2',
    icon: <Box size={32} color="#00d2ff" />,
    description: (
      <>
        Prefer object storage over OCI? Use the interactive wizard to automatically provision a Cloudflare R2 bucket and an edge Worker in seconds, mapping NARs directly to object keys.
      </>
    ),
  },
  {
    title: 'Zero-Config GHCR',
    icon: <Settings size={32} color="#00d2ff" />,
    description: (
      <>
        Seamlessly push your Nix cache to GitHub Container Registry or GitLab Registry using your standard Personal Access Token and GitHub Actions. No external hosting required.
      </>
    ),
  },
  {
    title: 'O(1) Manifest Lookups',
    icon: <Terminal size={32} color="#00d2ff" />,
    description: (
      <>
        By enforcing a strict Nix Hash tagging schema, Aeroflare fetches metadata instantly. When Nix requests <code>&lt;hash&gt;.narinfo</code>, the proxy pulls <code>registry/repo:&lt;hash&gt;</code> directly.
      </>
    ),
  },
  {
    title: 'Secure OS Keychain',
    icon: <Shield size={32} color="#00d2ff" />,
    description: (
      <>
        Tokens are never stored in plain text YAML. Aeroflare integrates natively with your OS keychain to keep your Cloudflare API keys and GitHub secrets safely encrypted.
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

function DeepDiveSection() {
  return (
    <div className="container" style={{ padding: '6rem 0', borderTop: '1px solid rgba(255,255,255,0.05)' }}>
      <div className="row">
        <div className="col col--12 text--center" style={{ marginBottom: '4rem' }}>
          <h2 style={{ fontSize: '2.5rem', fontWeight: 700, color: '#fff' }}>How It Works Under the Hood</h2>
          <p style={{ fontSize: '1.2rem', color: '#a3aed2', maxWidth: '800px', margin: '0 auto' }}>
            Aeroflare doesn't just proxy requests; it completely remaps the Nix protocol to container standards.
          </p>
        </div>
      </div>
      <div className="row" style={{ alignItems: 'center', marginBottom: '4rem' }}>
        <div className="col col--6">
          <h3 style={{ fontSize: '1.8rem', color: '#fff' }}><HardDrive size={24} style={{ marginRight: '10px', verticalAlign: 'middle', color: '#00d2ff' }}/> OCI Manifest Mapping</h3>
          <p style={{ color: '#a3aed2', fontSize: '1.1rem', lineHeight: 1.6 }}>
            Traditional Nix binary caches rely on a database or flat files to store <code>.narinfo</code> metadata. Aeroflare serializes this metadata (FileHash, StorePath, System, Sig) directly into the <strong>OCI Image Manifest Annotations</strong> (<code>vnd.aeroflare.nar.*</code>). The proxy then reconstructs the <code>.narinfo</code> on-the-fly when requested by the Nix daemon.
          </p>
        </div>
        <div className="col col--6">
          <div className="terminal-block" style={{ fontSize: '0.85rem' }}>
            <span style={{ color: '#00d2ff' }}>// Example OCI Manifest Annotation</span><br />
            &#123;<br />
            &nbsp;&nbsp;<span style={{ color: '#e5e7eb' }}>"vnd.aeroflare.nar.storepath"</span>: <span style={{ color: '#a8ff60' }}>"/nix/store/..."</span>,<br />
            &nbsp;&nbsp;<span style={{ color: '#e5e7eb' }}>"vnd.aeroflare.nar.compression"</span>: <span style={{ color: '#a8ff60' }}>"zstd"</span>,<br />
            &nbsp;&nbsp;<span style={{ color: '#e5e7eb' }}>"vnd.aeroflare.nar.narhash"</span>: <span style={{ color: '#a8ff60' }}>"sha256:..."</span><br />
            &#125;
          </div>
        </div>
      </div>
      <div className="row" style={{ alignItems: 'center' }}>
        <div className="col col--6">
          <div className="glass-card" style={{ padding: '2rem', textAlign: 'center' }}>
            <Cpu size={48} color="#00d2ff" style={{ marginBottom: '1rem' }} />
            <GitMerge size={48} color="#3a7bd5" style={{ marginLeft: '-15px', marginBottom: '1rem' }} />
            <h4 style={{ color: '#fff', fontSize: '1.2rem' }}>Ephemeral Proxy Substitution</h4>
          </div>
        </div>
        <div className="col col--6">
          <h3 style={{ fontSize: '1.8rem', color: '#fff' }}>Execution Wrapping</h3>
          <p style={{ color: '#a3aed2', fontSize: '1.1rem', lineHeight: 1.6 }}>
            The <code>aeroflare run</code> command wraps your standard <code>nix build</code>. It spawns a local HTTP server on a random port, injects <code>--option extra-substituters http://127.0.0.1:port</code> into your Nix invocation, and listens for requests. When the build finishes, it automatically identifies the new store paths and pushes them to your upstream cache concurrently.
          </p>
        </div>
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
      <main style={{ backgroundColor: '#0b0f19' }}>
        <div className="container" style={{ padding: '4rem 0' }}>
          <div className="row">
            {FeatureList.map((props, idx) => (
              <Feature key={idx} {...props} />
            ))}
          </div>
        </div>
        <DeepDiveSection />
      </main>
    </Layout>
  );
}
