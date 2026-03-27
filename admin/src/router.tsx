import { createBrowserRouter } from "react-router";

export const router = createBrowserRouter([
  {
    path: "/",
    lazy: () =>
      import("@/app/dashboard/pages/dashboard-page").then(
        ({ DashboardPage }) => ({
          Component: DashboardPage,
        }),
      ),
  },
]);
