import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useForm, useWatch } from "react-hook-form";

import { Button } from "@/components/ui/button";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import type { ApiResponse, User } from "@/lib/types";

interface FormValues {
  email: string;
  name: string;
  password?: string;
  role: string;
}

const roleOptions = {
  super_admin: [
    { value: "super_admin", label: "Super Admin" },
    { value: "admin", label: "Admin" },
    { value: "user", label: "User" },
  ],
  admin: [{ value: "user", label: "User" }],
};

export const UserFormSheet = ({
  mode,
  userId,
  currentUserRole,
  onClose,
}: {
  mode: "create" | "edit";
  userId?: string;
  currentUserRole: string;
  onClose: () => void;
}) => {
  const queryClient = useQueryClient();

  const detail = useQuery({
    queryKey: queryKeys.users.detail(userId ?? ""),
    queryFn: () =>
      api<ApiResponse<User>>(`/api/admin/users/${userId}`).then((r) => r.data),
    enabled: mode === "edit" && !!userId,
  });

  const isReady = mode === "create" || detail.isSuccess;
  const user = detail.data;
  const roles =
    roleOptions[currentUserRole as keyof typeof roleOptions] ??
    roleOptions.admin;

  return (
    <Sheet open onOpenChange={(open) => !open && onClose()}>
      <SheetContent side="right" className="sm:max-w-md">
        <SheetHeader>
          <SheetTitle>
            {mode === "create" ? "Create User" : "Edit User"}
          </SheetTitle>
          <SheetDescription>
            {mode === "create"
              ? "Add a new user to the system."
              : "Update user details."}
          </SheetDescription>
        </SheetHeader>

        {isReady && (
          <UserForm
            key={userId ?? "create"}
            mode={mode}
            userId={userId}
            defaultValues={
              user
                ? { email: user.email, name: user.name, role: user.role_slug }
                : { email: "", name: "", role: roles[roles.length - 1].value }
            }
            roles={roles}
            onSuccess={() => {
              queryClient.invalidateQueries({
                queryKey: queryKeys.users.all,
              });
              onClose();
            }}
          />
        )}
      </SheetContent>
    </Sheet>
  );
};

const UserForm = ({
  mode,
  userId,
  defaultValues,
  roles,
  onSuccess,
}: {
  mode: "create" | "edit";
  userId?: string;
  defaultValues: Partial<FormValues>;
  roles: { value: string; label: string }[];
  onSuccess: () => void;
}) => {
  const {
    register,
    handleSubmit,
    setValue,
    setError,
    control,
    formState: { errors },
  } = useForm<FormValues>({ defaultValues });

  const mutation = useMutation({
    mutationFn: (values: FormValues) => {
      if (mode === "create") {
        return api("/api/admin/users", {
          method: "POST",
          body: JSON.stringify(values),
        });
      }
      return api(`/api/admin/users/${userId}`, {
        method: "PATCH",
        body: JSON.stringify({
          name: values.name,
          email: values.email,
          role: values.role,
        }),
      });
    },
    onSuccess,
    onError: (error: Error) => {
      setError("root", { message: error.message });
    },
  });

  const selectedRole =
    useWatch({ control, name: "role" }) ?? roles[roles.length - 1].value;

  return (
    <form
      onSubmit={handleSubmit((v) => mutation.mutate(v))}
      className="flex flex-1 flex-col"
    >
      <div className="flex-1 overflow-y-auto px-4 py-4">
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="uf-name">Name</FieldLabel>
            <Input
              id="uf-name"
              aria-invalid={!!errors.name}
              {...register("name", { required: "Name is required" })}
            />
            <FieldError errors={[errors.name]} />
          </Field>
          <Field>
            <FieldLabel htmlFor="uf-email">Email</FieldLabel>
            <Input
              id="uf-email"
              type="email"
              aria-invalid={!!errors.email}
              {...register("email", {
                required: "Email is required",
                pattern: {
                  value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                  message: "Enter a valid email address",
                },
              })}
            />
            <FieldError errors={[errors.email]} />
          </Field>
          {mode === "create" && (
            <Field>
              <FieldLabel htmlFor="uf-password">Password</FieldLabel>
              <Input
                id="uf-password"
                type="password"
                aria-invalid={!!errors.password}
                {...register("password", {
                  required: "Password is required",
                  minLength: {
                    value: 8,
                    message: "Password must be at least 8 characters",
                  },
                })}
              />
              <FieldError errors={[errors.password]} />
            </Field>
          )}
          <Field>
            <FieldLabel>Role</FieldLabel>
            <Select
              value={selectedRole}
              onValueChange={(v) => setValue("role", v)}
            >
              <SelectTrigger aria-invalid={!!errors.role}>
                <SelectValue placeholder="Select role" />
              </SelectTrigger>
              <SelectContent>
                {roles.map((r) => (
                  <SelectItem key={r.value} value={r.value}>
                    {r.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <FieldError errors={[errors.role]} />
          </Field>
          <FieldError errors={[errors.root]} />
        </FieldGroup>
      </div>

      <SheetFooter className="px-4 py-4">
        <Button type="submit" disabled={mutation.isPending} className="w-full">
          {mutation.isPending && <Loader2 className="animate-spin" />}
          {mode === "create"
            ? mutation.isPending
              ? "Creating\u2026"
              : "Create User"
            : mutation.isPending
              ? "Saving\u2026"
              : "Save Changes"}
        </Button>
      </SheetFooter>
    </form>
  );
};
