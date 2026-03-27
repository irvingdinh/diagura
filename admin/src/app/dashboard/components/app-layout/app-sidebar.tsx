import type { ComponentProps } from "react";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar.tsx";

export const AppSidebar = ({ ...props }: ComponentProps<typeof Sidebar>) => {
  return (
    <Sidebar {...props}>
      <SidebarHeader>&nbsp;</SidebarHeader>
      <SidebarContent>&nbsp;</SidebarContent>
      <SidebarRail />
      <SidebarFooter>&nbsp;</SidebarFooter>
    </Sidebar>
  );
};
