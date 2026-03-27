import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import { queryKeys } from "@/lib/query-keys";

interface PasswordFormValues {
  current_password: string;
  new_password: string;
}

export const ChangePasswordForm = ({
  forceChange,
}: {
  forceChange: boolean;
}) => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const {
    register,
    handleSubmit,
    setError,
    reset,
    formState: { errors },
  } = useForm<PasswordFormValues>();

  const mutation = useMutation({
    mutationFn: (values: PasswordFormValues) =>
      api("/api/admin/profile/password", {
        method: "PUT",
        body: JSON.stringify(values),
      }),
    onSuccess: () => {
      reset();
      queryClient.invalidateQueries({ queryKey: queryKeys.session });
      if (forceChange) {
        navigate("/admin");
      }
    },
    onError: (error: Error) => {
      setError("root", { message: error.message });
    },
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Change Password</CardTitle>
        <CardDescription>
          Update your password. All other sessions will be logged out.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit((v) => mutation.mutate(v))}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="current_password">
                Current Password
              </FieldLabel>
              <Input
                id="current_password"
                type="password"
                aria-invalid={!!errors.current_password}
                {...register("current_password", {
                  required: "Current password is required",
                })}
              />
              <FieldError errors={[errors.current_password]} />
            </Field>
            <Field>
              <FieldLabel htmlFor="new_password">New Password</FieldLabel>
              <Input
                id="new_password"
                type="password"
                aria-invalid={!!errors.new_password || !!errors.root}
                {...register("new_password", {
                  required: "New password is required",
                  minLength: {
                    value: 8,
                    message: "Password must be at least 8 characters",
                  },
                })}
              />
              <FieldError errors={[errors.new_password, errors.root]} />
            </Field>
            <Field>
              <Button type="submit" disabled={mutation.isPending}>
                {mutation.isPending && <Loader2 className="animate-spin" />}
                {mutation.isPending
                  ? "Changing password\u2026"
                  : "Change Password"}
              </Button>
            </Field>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  );
};
