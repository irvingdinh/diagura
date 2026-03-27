import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useForm } from "react-hook-form";

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
import type { SessionUser } from "@/lib/types";

interface NameFormValues {
  name: string;
}

export const UpdateNameForm = ({ user }: { user: SessionUser }) => {
  const queryClient = useQueryClient();

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<NameFormValues>({ defaultValues: { name: user.name } });

  const mutation = useMutation({
    mutationFn: (values: NameFormValues) =>
      api("/api/admin/profile", {
        method: "PATCH",
        body: JSON.stringify(values),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.session });
    },
    onError: (error: Error) => {
      setError("root", { message: error.message });
    },
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Profile</CardTitle>
        <CardDescription>Update your display name.</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit((v) => mutation.mutate(v))}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="name">Name</FieldLabel>
              <Input
                id="name"
                aria-invalid={!!errors.name}
                {...register("name", { required: "Name is required" })}
              />
              <FieldError errors={[errors.name, errors.root]} />
            </Field>
            <Field>
              <Button type="submit" disabled={mutation.isPending}>
                {mutation.isPending && <Loader2 className="animate-spin" />}
                {mutation.isPending ? "Saving\u2026" : "Save"}
              </Button>
            </Field>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  );
};
