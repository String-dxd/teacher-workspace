import React from 'react';
import { Outlet } from 'react-router';

import { AppSidebar } from '~/components/Sidebar';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '~/components/ui/sidebar';
import { WelcomeModal } from '~/components/WelcomeModal';

export default function RootLayout() {
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="tw:flex tw:h-14 tw:items-center tw:px-4 tw:md:hidden">
          <SidebarTrigger />
        </header>
        <React.Suspense fallback={null}>
          <Outlet />
        </React.Suspense>
      </SidebarInset>
      <WelcomeModal />
    </SidebarProvider>
  );
}
