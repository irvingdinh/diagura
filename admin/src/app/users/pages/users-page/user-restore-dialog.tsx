import { useMutation, useQueryClient } from "@tanstack/react-query";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { api } from "@/lib/api";
import { queryKeys } from "@/lib/query-keys";
import type { User } from "@/lib/types";

export const UserRestoreDialog = ({
  user,
  onClose,
}: {
  user: User;
  onClose: () => void;
}) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () =>
      api(`/api/admin/users/${user.id}/restore`, { method: "POST" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      onClose();
    },
  });

  return (
    <AlertDialog open onOpenChange={(open) => !open && onClose()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Restore User</AlertDialogTitle>
          <AlertDialogDescription>
            Restore {user.name}? They will be required to change their password
            on next login.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={() => mutation.mutate()}
            disabled={mutation.isPending}
          >
            {mutation.isPending ? "Restoring\u2026" : "Restore"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};
