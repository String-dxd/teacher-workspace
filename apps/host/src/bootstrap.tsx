import { registerRemotes } from '@module-federation/enhanced/runtime';
import React from 'react';
import { createRoot } from 'react-dom/client';

import './App.css';
import App from './App';

interface RuntimeRemote {
  name: string;
  entry: string;
}

interface RuntimeConfig {
  remotes: RuntimeRemote[];
}

// A page served without the block, such as one loaded straight from the rsbuild
// dev server, runs with no remotes and the routes that need one show their fallback.
function readRuntimeConfig(): RuntimeConfig {
  try {
    const source = document.getElementById('runtime-config')?.textContent ?? '';
    const config = (JSON.parse(source) ?? {}) as Partial<RuntimeConfig>;
    return { remotes: config.remotes ?? [] };
  } catch (error: unknown) {
    console.error('Could not read the runtime config from the page; no remotes registered', error);
    return { remotes: [] };
  }
}

const container = document.getElementById('root');
if (!container) throw new Error('Root element #root not found');

// Registered before the first render, so a route that lazy-loads a remote
// always finds its entry.
registerRemotes(readRuntimeConfig().remotes);

createRoot(container).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
