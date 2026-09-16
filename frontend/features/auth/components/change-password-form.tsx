"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useTranslations } from "next-intl";
import { z } from "zod";
import { toast } from "sonner";

import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useRouter } from "@/i18n/navigation";
import { useChangePassword } from "@/hooks/use-auth";
import { ApiError } from "@/lib/api-error";

const schema = z
  .object({
    currentPassword: z.string().min(1),
    newPassword: z.string().min(8).max(72),
    confirmPassword: z.string().min(1),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    path: ["confirmPassword"],
    message: "passwordsDontMatch",
  });

export function ChangePasswordForm() {
  const t = useTranslations("auth");
  const router = useRouter();
  const changePassword = useChangePassword();

  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { currentPassword: "", newPassword: "", confirmPassword: "" },
  });

  async function onSubmit(values: z.infer<typeof schema>) {
    try {
      await changePassword.mutateAsync({
        currentPassword: values.currentPassword,
        newPassword: values.newPassword,
      });
      toast.success(t("changePasswordSuccess"));
      router.push("/login");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        form.setError("currentPassword", { message: t("currentPasswordIncorrect") });
      } else {
        toast.error(err instanceof ApiError ? err.message : t("changePasswordFailed"));
      }
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="currentPassword"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("currentPassword")}</FormLabel>
              <FormControl>
                <Input type="password" autoComplete="current-password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
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
        <FormField
          control={form.control}
          name="confirmPassword"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("confirmNewPassword")}</FormLabel>
              <FormControl>
                <Input type="password" autoComplete="new-password" {...field} />
              </FormControl>
              {form.formState.errors.confirmPassword?.message === "passwordsDontMatch" ? (
                <p className="text-sm text-destructive">{t("passwordsDontMatch")}</p>
              ) : (
                <FormMessage />
              )}
            </FormItem>
          )}
        />
        <Button type="submit" className="w-full" disabled={changePassword.isPending}>
          {t("changePassword")}
        </Button>
      </form>
    </Form>
  );
}
