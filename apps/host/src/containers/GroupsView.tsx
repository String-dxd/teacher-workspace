import { loadRemote, registerRemotes } from '@module-federation/enhanced/runtime';
import React from 'react';

import { ErrorBoundary } from '~/components/ErrorBoundary';
import { RemoteLoadFallbackView } from '~/containers/RemoteLoadFallbackView';

registerRemotes([
  {
    name: 'pg',
    entry: 'https://d390008ekba73v.cloudfront.net/mf-manifest.json',
  },
]);

const RemoteApp = React.lazy(async () => {
  const module = await loadRemote<{ default: React.ComponentType }>('pg/Groups');

  if (!module) {
    throw new Error('Failed to load remote module');
  }

  return module;
});

export default function GroupsView() {
  return (
    <ErrorBoundary fallback={<RemoteLoadFallbackView onRetry={() => window.location.reload()} />}>
      <React.Suspense fallback={null}>
        <RemoteApp />
      </React.Suspense>
    </ErrorBoundary>
  );
}
