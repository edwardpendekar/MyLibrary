"use client";

import { useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useTranslations } from "next-intl";
import { z } from "zod";
import { toast } from "sonner";

import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useForgotPassword } from "@/hooks/use-auth";
import { ApiError } from "@/lib/api-error";

const schema = z.object({
  email: z.string().email(),
});

export function ForgotPasswordForm() {
  const t = useTranslations("auth");
  const tCommon = useTranslations("common");
  const [sent, setSent] = useState(false);
  const forgotPassword = useForgotPassword();

  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { email: "" },
  });

  async function onSubmit(values: z.infer<typeof schema>) {
    try {
      await forgotPassword.mutateAsync(values.email);
      setSent(true);
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : tCommon("error"));
    }
  }

  if (sent) {
    return <p className="text-sm text-muted-foreground">{t("resetLinkSent")}</p>;
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <p className="text-sm text-muted-foreground">{t("forgotPasswordHint")}</p>
        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("email")}</FormLabel>
              <FormControl>
                <Input type="email" autoComplete="email" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit" className="w-full" disabled={forgotPassword.isPending}>
          {t("sendResetLink")}
        </Button>
      </form>
    </Form>
  );
}
