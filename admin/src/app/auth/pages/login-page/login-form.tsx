import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import type { ComponentProps } from "react";
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
import { queryKeys } from "@/lib/query-keys";
import type { ApiResponse, SessionUser } from "@/lib/types";
import { cn } from "@/lib/utils";

interface LoginFormValues {
  email: string;
  password: string;
}

async function login(
  values: LoginFormValues,
): Promise<ApiResponse<SessionUser>> {
  const res = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(values),
  });

  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Something went wrong");
  }

  return res.json();
}

export const LoginForm = ({ className, ...props }: ComponentProps<"div">) => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<LoginFormValues>();

  const mutation = useMutation({
    mutationFn: login,
    onSuccess: (response) => {
      queryClient.setQueryData(queryKeys.session, response.data);

      if (response.data.force_password_change) {
        navigate("/admin/profile");
      } else {
        navigate("/admin");
      }
    },
    onError: (error: Error) => {
      setError("root", { message: error.message });
    },
  });

  const onSubmit = handleSubmit((values) => mutation.mutate(values));

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Login to your account</CardTitle>
          <CardDescription>
            Enter your email below to login to your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit}>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="email">Email</FieldLabel>
                <Input
                  id="email"
                  type="email"
                  placeholder="m@example.com"
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
              <Field>
                <div className="flex items-center">
                  <FieldLabel htmlFor="password">Password</FieldLabel>
                </div>
                <Input
                  id="password"
                  type="password"
                  aria-invalid={!!errors.password || !!errors.root}
                  {...register("password", {
                    required: "Password is required",
                  })}
                />
                <FieldError errors={[errors.password, errors.root]} />
              </Field>
              <Field>
                <Button type="submit" disabled={mutation.isPending}>
                  {mutation.isPending && <Loader2 className="animate-spin" />}
                  {mutation.isPending ? "Logging in\u2026" : "Login"}
                </Button>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
    </div>
  );
};
