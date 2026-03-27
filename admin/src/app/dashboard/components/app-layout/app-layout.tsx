import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router";

import { AppSidebar } from "@/app/dashboard/components/app-layout/app-sidebar.tsx";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar.tsx";
import { Skeleton } from "@/components/ui/skeleton.tsx";
import { useSession } from "@/hooks/use-session.ts";

export const AppLayout = ({ children }: { children: ReactNode }) => {
  const session = useSession();
  const { pathname } = useLocation();

  if (session.isLoading) {
    return (
      <SidebarProvider>
        <div className="flex h-screen w-64 flex-col gap-4 border-r p-4">
          <Skeleton className="h-8 w-32" />
          <Skeleton className="h-6 w-full" />
          <Skeleton className="h-6 w-full" />
        </div>
        <SidebarInset>
          <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <Skeleton className="h-6 w-6" />
          </header>
          <main className="flex flex-1 flex-col gap-4 p-4">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-64 w-full" />
          </main>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  if (session.isError || !session.data) {
    return <Navigate to="/admin/login" replace />;
  }

  if (session.data.force_password_change && pathname !== "/admin/profile") {
    return <Navigate to="/admin/profile" replace />;
  }

  return (
    <SidebarProvider>
      <AppSidebar user={session.data} />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
        </header>
        <main className="flex flex-1 flex-col gap-4 p-4">{children}</main>
      </SidebarInset>
    </SidebarProvider>
  );
};
