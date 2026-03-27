import {
  KeyRound,
  MoreHorizontal,
  Pencil,
  RotateCcw,
  Trash2,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { SessionUser, User } from "@/lib/types";

function canManage(actorRole: string, targetRoleSlug: string): boolean {
  if (actorRole === "super_admin") return true;
  return actorRole === "admin" && targetRoleSlug === "user";
}

function formatDate(dateStr: string): string {
  return new Date(dateStr + "Z").toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export const UsersTable = ({
  data,
  isLoading,
  currentUser,
  onEdit,
  onSetPassword,
  onDelete,
  onRestore,
}: {
  data: User[];
  isLoading: boolean;
  currentUser: SessionUser;
  onEdit: (user: User) => void;
  onSetPassword: (user: User) => void;
  onDelete: (user: User) => void;
  onRestore: (user: User) => void;
}) => {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Role</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
            <TableHead className="w-12" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading &&
            Array.from({ length: 5 }).map((_, i) => (
              <TableRow key={i}>
                <TableCell>
                  <Skeleton className="h-4 w-24" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-4 w-36" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-5 w-16" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-5 w-14" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-4 w-20" />
                </TableCell>
                <TableCell />
              </TableRow>
            ))}

          {!isLoading && data.length === 0 && (
            <TableRow>
              <TableCell
                colSpan={6}
                className="text-muted-foreground h-24 text-center"
              >
                No users found.
              </TableCell>
            </TableRow>
          )}

          {!isLoading &&
            data.map((user) => {
              const manageable = canManage(currentUser.role, user.role_slug);
              const isSelf = user.id === currentUser.id;
              const isDeleted = !!user.deleted_at;

              return (
                <TableRow key={user.id}>
                  <TableCell className="font-medium">{user.name}</TableCell>
                  <TableCell>{user.email}</TableCell>
                  <TableCell>
                    <Badge variant="outline">{user.role_name}</Badge>
                  </TableCell>
                  <TableCell>
                    {isDeleted ? (
                      <Badge variant="destructive">Deleted</Badge>
                    ) : (
                      <Badge variant="secondary">Active</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {formatDate(user.created_at)}
                  </TableCell>
                  <TableCell>
                    {manageable && !isSelf && (
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon-xs">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          {!isDeleted && (
                            <>
                              <DropdownMenuItem onClick={() => onEdit(user)}>
                                <Pencil className="mr-2 h-4 w-4" />
                                Edit
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => onSetPassword(user)}
                              >
                                <KeyRound className="mr-2 h-4 w-4" />
                                Set Password
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                className="text-destructive"
                                onClick={() => onDelete(user)}
                              >
                                <Trash2 className="mr-2 h-4 w-4" />
                                Delete
                              </DropdownMenuItem>
                            </>
                          )}
                          {isDeleted && (
                            <DropdownMenuItem onClick={() => onRestore(user)}>
                              <RotateCcw className="mr-2 h-4 w-4" />
                              Restore
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    )}
                  </TableCell>
                </TableRow>
              );
            })}
        </TableBody>
      </Table>
    </div>
  );
};
