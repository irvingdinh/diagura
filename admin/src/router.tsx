import type { ComponentType } from "react";
import { createBrowserRouter, Navigate, Outlet } from "react-router";

import { AppLayout } from "@/app/dashboard/components/app-layout";

function lazy(factory: () => Promise<Record<string, ComponentType>>) {
  return {
    lazy: () =>
      factory().then((mod) => ({
        Component: Object.values(mod)[0] as ComponentType,
      })),
  };
}

export const router = createBrowserRouter([
  {
    path: "/admin/login",
    hydrateFallbackElement: null,
    ...lazy(() => import("@/app/auth/pages/login-page")),
  },
  {
    path: "/",
    Component: () => <Navigate to="/admin" replace />,
  },
  {
    path: "/admin",
    hydrateFallbackElement: null,
    Component: () => (
      <AppLayout>
        <Outlet />
      </AppLayout>
    ),
    children: [
      {
        index: true,
        ...lazy(() => import("@/app/dashboard/pages/dashboard-page")),
      },
      {
        path: "profile",
        ...lazy(() => import("@/app/profile/pages/profile-page")),
      },
      {
        path: "users",
        ...lazy(() => import("@/app/users/pages/users-page")),
      },
      {
        path: "logs",
        ...lazy(() => import("@/app/logs/pages/logs-page")),
      },
    ],
  },
]);
