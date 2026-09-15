import { getTranslations } from "next-intl/server";

import { Link } from "@/i18n/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LoginForm } from "@/features/auth/components/login-form";

export default async function LoginPage() {
  const t = await getTranslations("auth");

  return (
    <div className="mx-auto flex max-w-sm flex-col justify-center px-4 py-16">
      <Card>
        <CardHeader>
          <CardTitle>{t("loginTitle")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <LoginForm />
          <p className="text-center text-sm text-muted-foreground">
            {t("noAccount")}{" "}
            <Link href="/register" className="underline">
              {t("submitRegister")}
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
