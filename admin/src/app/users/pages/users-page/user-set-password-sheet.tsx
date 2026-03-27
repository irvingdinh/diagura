import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { api } from "@/lib/api";
import { queryKeys } from "@/lib/query-keys";

interface PasswordFormValues {
  password: string;
}

export const UserSetPasswordSheet = ({
  userId,
  userName,
  onClose,
}: {
  userId: string;
  userName: string;
  onClose: () => void;
}) => {
  const queryClient = useQueryClient();

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<PasswordFormValues>();

  const mutation = useMutation({
    mutationFn: (values: PasswordFormValues) =>
      api(`/api/admin/users/${userId}/password`, {
        method: "PUT",
        body: JSON.stringify(values),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      onClose();
    },
    onError: (error: Error) => {
      setError("root", { message: error.message });
    },
  });

  return (
    <Sheet open onOpenChange={(open) => !open && onClose()}>
      <SheetContent side="right" className="sm:max-w-md">
        <SheetHeader>
          <SheetTitle>Set Password</SheetTitle>
          <SheetDescription>
            Set a temporary password for {userName}. They will be required to
            change it on next login. All their current sessions will be logged
            out.
          </SheetDescription>
        </SheetHeader>

        <form
          onSubmit={handleSubmit((v) => mutation.mutate(v))}
          className="flex flex-1 flex-col"
        >
          <div className="flex-1 px-4 py-4">
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="sp-password">New Password</FieldLabel>
                <Input
                  id="sp-password"
                  type="password"
                  aria-invalid={!!errors.password || !!errors.root}
                  {...register("password", {
                    required: "Password is required",
                    minLength: {
                      value: 8,
                      message: "Password must be at least 8 characters",
                    },
                  })}
                />
                <FieldError errors={[errors.password, errors.root]} />
              </Field>
            </FieldGroup>
          </div>

          <SheetFooter className="px-4 py-4">
            <Button
              type="submit"
              disabled={mutation.isPending}
              className="w-full"
            >
              {mutation.isPending && <Loader2 className="animate-spin" />}
              {mutation.isPending ? "Setting password\u2026" : "Set Password"}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
};
