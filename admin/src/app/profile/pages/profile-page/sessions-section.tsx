import { useMutation } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { api } from "@/lib/api";

export const SessionsSection = () => {
  const mutation = useMutation({
    mutationFn: () =>
      api("/api/admin/profile/sessions/logout", { method: "POST" }),
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Sessions</CardTitle>
        <CardDescription>
          Log out all other sessions across all your devices.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Button
          variant="destructive"
          disabled={mutation.isPending || mutation.isSuccess}
          onClick={() => mutation.mutate()}
        >
          {mutation.isPending && <Loader2 className="animate-spin" />}
          {mutation.isSuccess
            ? "Sessions logged out"
            : mutation.isPending
              ? "Logging out\u2026"
              : "Log Out Other Sessions"}
        </Button>
      </CardContent>
    </Card>
  );
};
