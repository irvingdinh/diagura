import type { ComponentType } from "react";
import { createBrowserRouter, Outlet } from "react-router";

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
    path: "/",
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
    ],
  },
]);
