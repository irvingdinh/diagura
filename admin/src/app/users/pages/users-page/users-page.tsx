import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { useSession } from "@/hooks/use-session";
import { api } from "@/lib/api";
import { queryKeys } from "@/lib/query-keys";
import type { ApiListResponse, User, UserListParams } from "@/lib/types";

import { UserDeleteDialog } from "./user-delete-dialog";
import { UserFormSheet } from "./user-form-sheet";
import { UserRestoreDialog } from "./user-restore-dialog";
import { UserSetPasswordSheet } from "./user-set-password-sheet";
import { UsersTable } from "./users-table";
import { UsersToolbar } from "./users-toolbar";

type SheetState =
  | { mode: "create" }
  | { mode: "edit"; userId: string }
  | { mode: "password"; userId: string; userName: string }
  | null;

type DialogState =
  | { mode: "delete"; user: User }
  | { mode: "restore"; user: User }
  | null;

export const UsersPage = () => {
  const { data: session } = useSession();
  const [params, setParams] = useState<UserListParams>({
    page: 1,
    per_page: 20,
    status: "active",
  });
  const [sheet, setSheet] = useState<SheetState>(null);
  const [dialog, setDialog] = useState<DialogState>(null);

  const query = useQuery({
    queryKey: queryKeys.users.list(params as Record<string, unknown>),
    queryFn: () => {
      const qs = new URLSearchParams();
      if (params.page) qs.set("page", String(params.page));
      if (params.per_page) qs.set("per_page", String(params.per_page));
      if (params.search) qs.set("search", params.search);
      if (params.role) qs.set("role", params.role);
      if (params.status) qs.set("status", params.status);
      return api<ApiListResponse<User>>(`/api/admin/users?${qs}`);
    },
    placeholderData: keepPreviousData,
  });

  if (!session) return null;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Users</h1>
        <Button onClick={() => setSheet({ mode: "create" })}>
          <Plus className="mr-1 h-4 w-4" />
          New User
        </Button>
      </div>

      <UsersToolbar
        params={params}
        onChange={(next) => setParams({ ...next, page: 1 })}
      />

      <UsersTable
        data={query.data?.data ?? []}
        isLoading={query.isLoading}
        currentUser={session}
        onEdit={(user) => setSheet({ mode: "edit", userId: user.id })}
        onSetPassword={(user) =>
          setSheet({ mode: "password", userId: user.id, userName: user.name })
        }
        onDelete={(user) => setDialog({ mode: "delete", user })}
        onRestore={(user) => setDialog({ mode: "restore", user })}
      />

      {query.data?.meta && query.data.meta.total_pages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground text-sm">
            Showing {(query.data.meta.page - 1) * query.data.meta.per_page + 1}
            &ndash;
            {Math.min(
              query.data.meta.page * query.data.meta.per_page,
              query.data.meta.total,
            )}{" "}
            of {query.data.meta.total}
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={query.data.meta.page <= 1}
              onClick={() =>
                setParams((p) => ({ ...p, page: (p.page ?? 1) - 1 }))
              }
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={query.data.meta.page >= query.data.meta.total_pages}
              onClick={() =>
                setParams((p) => ({ ...p, page: (p.page ?? 1) + 1 }))
              }
            >
              Next
            </Button>
          </div>
        </div>
      )}

      {sheet?.mode === "create" && (
        <UserFormSheet
          mode="create"
          currentUserRole={session.role}
          onClose={() => setSheet(null)}
        />
      )}
      {sheet?.mode === "edit" && (
        <UserFormSheet
          mode="edit"
          userId={sheet.userId}
          currentUserRole={session.role}
          onClose={() => setSheet(null)}
        />
      )}
      {sheet?.mode === "password" && (
        <UserSetPasswordSheet
          userId={sheet.userId}
          userName={sheet.userName}
          onClose={() => setSheet(null)}
        />
      )}
      {dialog?.mode === "delete" && (
        <UserDeleteDialog user={dialog.user} onClose={() => setDialog(null)} />
      )}
      {dialog?.mode === "restore" && (
        <UserRestoreDialog user={dialog.user} onClose={() => setDialog(null)} />
      )}
    </div>
  );
};
