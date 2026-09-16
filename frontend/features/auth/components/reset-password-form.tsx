"use client";

import { useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useTranslations } from "next-intl";
import { z } from "zod";
import { toast } from "sonner";

import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button, buttonVariants } from "@/components/ui/button";
import { Link, useRouter } from "@/i18n/navigation";
import { useResetPassword } from "@/hooks/use-auth";
import { ApiError } from "@/lib/api-error";

const schema = z.object({
  newPassword: z.string().min(8).max(72),
});

export function ResetPasswordForm({ token }: { token: string }) {
  const t = useTranslations("auth");
  const router = useRouter();
  const [invalid, setInvalid] = useState(false);
  const resetPassword = useResetPassword();

  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { newPassword: "" },
  });

  async function onSubmit(values: z.infer<typeof schema>) {
    try {
      await resetPassword.mutateAsync({ token, newPassword: values.newPassword });
      toast.success(t("resetPasswordSuccess"));
      router.push("/login");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setInvalid(true);
      } else {
        toast.error(err instanceof ApiError ? err.message : t("resetPasswordInvalid"));
      }
    }
  }

  if (invalid) {
    return (
      <div className="space-y-4">
        <p className="text-sm text-destructive">{t("resetPasswordInvalid")}</p>
        <Link href="/forgot-password" className={buttonVariants({ className: "w-full" })}>
          {t("requestNewLink")}
        </Link>
      </div>
    );
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="newPassword"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("newPassword")}</FormLabel>
              <FormControl>
                <Input type="password" autoComplete="new-password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit" className="w-full" disabled={resetPassword.isPending}>
          {t("resetPassword")}
        </Button>
      </form>
    </Form>
  );
}
