import { getTranslations } from "next-intl/server";

import { Link } from "@/i18n/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { buttonVariants } from "@/components/ui/button";
import { ResetPasswordForm } from "@/features/auth/components/reset-password-form";

export default async function ResetPasswordPage({
  searchParams,
}: PageProps<"/[locale]/reset-password">) {
  const t = await getTranslations("auth");
  const { token } = await searchParams;
  const tokenValue = typeof token === "string" ? token : "";

  return (
    <div className="mx-auto flex max-w-sm flex-col justify-center px-4 py-16">
      <Card>
        <CardHeader>
          <CardTitle>{t("resetPasswordTitle")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {tokenValue ? (
            <ResetPasswordForm token={tokenValue} />
          ) : (
            <div className="space-y-4">
              <p className="text-sm text-destructive">{t("resetPasswordInvalid")}</p>
              <Link href="/forgot-password" className={buttonVariants({ className: "w-full" })}>
                {t("requestNewLink")}
              </Link>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
