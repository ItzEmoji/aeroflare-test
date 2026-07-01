import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import { motion } from 'framer-motion';

function HomepageHeader() {
  return (
    <header className="hero-gradient-bg" style={{ minHeight: 'calc(100vh - 60px)', display: 'flex', alignItems: 'center', justifyContent: 'center', textAlign: 'center' }}>
      <div className="container">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, ease: "easeOut" }}
        >
          <img src="/img/logo.svg" alt="Aeroflare Logo" style={{ width: '120px', marginBottom: '2.5rem' }} />
          <h1 className="hero__title" style={{ fontSize: '4.5rem', fontWeight: 700, marginBottom: '3rem', letterSpacing: '-0.02em', lineHeight: 1.1, color: '#fff' }}>
            The OCI-Nix-Binary-Cache <br /> <span className="gradient-text">written in Go</span>
          </h1>
          
          <motion.div 
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.6, delay: 0.3 }}
            style={{ maxWidth: '650px', margin: '0 auto' }}
          >
            <div className="terminal-block" style={{ padding: '1.5rem 2rem', fontSize: '1.1rem', textAlign: 'center' }}>
              <span className="terminal-prompt">$</span> <span style={{ color: '#fff' }}>nix run github:ItzEmoji/aeroflare -- init</span>
            </div>
          </motion.div>
        </motion.div>
      </div>
    </header>
  );
}

export default function Home(): JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="The OCI-Nix-Binary-Cache written in Go">
      <HomepageHeader />
    </Layout>
  );
}
