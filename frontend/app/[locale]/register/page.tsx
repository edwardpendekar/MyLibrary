import { getTranslations } from "next-intl/server";

import { Link } from "@/i18n/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { RegisterForm } from "@/features/auth/components/register-form";

export default async function RegisterPage() {
  const t = await getTranslations("auth");

  return (
    <div className="mx-auto flex max-w-sm flex-col justify-center px-4 py-16">
      <Card>
        <CardHeader>
          <CardTitle>{t("registerTitle")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <RegisterForm />
          <p className="text-center text-sm text-muted-foreground">
            {t("haveAccount")}{" "}
            <Link href="/login" className="underline">
              {t("submitLogin")}
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
