import { useEffect, useRef, useState } from "react";

import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { UserListParams } from "@/lib/types";

export const UsersToolbar = ({
  params,
  onChange,
}: {
  params: UserListParams;
  onChange: (params: UserListParams) => void;
}) => {
  const [searchInput, setSearchInput] = useState(params.search ?? "");
  const timerRef = useRef<ReturnType<typeof setTimeout>>(null);
  const paramsRef = useRef(params);
  useEffect(() => {
    paramsRef.current = params;
  });

  const handleSearch = (value: string) => {
    setSearchInput(value);
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      onChange({ ...paramsRef.current, search: value || undefined });
    }, 300);
  };

  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
      <Input
        placeholder="Search users..."
        value={searchInput}
        onChange={(e) => handleSearch(e.target.value)}
        className="sm:max-w-xs"
      />
      <Select
        value={params.role ?? "all"}
        onValueChange={(v) =>
          onChange({ ...params, role: v === "all" ? undefined : v })
        }
      >
        <SelectTrigger className="sm:w-40">
          <SelectValue placeholder="All Roles" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Roles</SelectItem>
          <SelectItem value="super_admin">Super Admin</SelectItem>
          <SelectItem value="admin">Admin</SelectItem>
          <SelectItem value="user">User</SelectItem>
        </SelectContent>
      </Select>
      <Select
        value={params.status ?? "active"}
        onValueChange={(v) =>
          onChange({ ...params, status: v as "active" | "deleted" })
        }
      >
        <SelectTrigger className="sm:w-32">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="active">Active</SelectItem>
          <SelectItem value="deleted">Deleted</SelectItem>
        </SelectContent>
      </Select>
    </div>
  );
};
