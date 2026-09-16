import { getTranslations } from "next-intl/server";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ChangePasswordForm } from "@/features/auth/components/change-password-form";

export default async function ChangePasswordPage() {
  const t = await getTranslations("auth");

  return (
    <div className="mx-auto flex max-w-sm flex-col justify-center px-4 py-16">
      <Card>
        <CardHeader>
          <CardTitle>{t("changePasswordTitle")}</CardTitle>
        </CardHeader>
        <CardContent>
          <ChangePasswordForm />
        </CardContent>
      </Card>
    </div>
  );
}
