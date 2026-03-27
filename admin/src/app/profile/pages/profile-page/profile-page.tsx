import { AlertCircle } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { useSession } from "@/hooks/use-session";

import { ChangePasswordForm } from "./change-password-form";
import { SessionsSection } from "./sessions-section";
import { UpdateNameForm } from "./update-name-form";

export const ProfilePage = () => {
  const { data: user } = useSession();

  if (!user) return null;

  return (
    <div className="mx-auto w-full max-w-2xl space-y-6">
      {user.force_password_change && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Password change required</AlertTitle>
          <AlertDescription>
            You must change your password before continuing.
          </AlertDescription>
        </Alert>
      )}

      <UpdateNameForm user={user} />
      <ChangePasswordForm forceChange={user.force_password_change} />
      <SessionsSection />
    </div>
  );
};
