import { loadRemote, registerRemotes } from '@module-federation/enhanced/runtime';
import React from 'react';

import { ErrorBoundary } from '~/components/ErrorBoundary';
import { RemoteLoadFallback } from '~/components/RemoteLoadFallback';

registerRemotes([
  {
    name: 'pg',
    entry: 'https://d390008ekba73v.cloudfront.net/mf-manifest.json',
  },
]);

const RemoteApp = React.lazy(async () => {
  const module = await loadRemote<{ default: React.ComponentType }>('pg/Posts');

  if (!module) {
    throw new Error('Failed to load remote module');
  }

  return module;
});

function Fallback() {
  return <RemoteLoadFallback onRetry={() => window.location.reload()} />;
}

export default function PostsView() {
  return (
    <ErrorBoundary fallback={<Fallback />}>
      <React.Suspense fallback={null}>
        <RemoteApp />
      </React.Suspense>
    </ErrorBoundary>
  );
}
