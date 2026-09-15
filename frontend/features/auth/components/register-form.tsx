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
import { useLogin, useRegister } from "@/hooks/use-auth";
import { ApiError } from "@/lib/api-error";

const schema = z.object({
  name: z.string().min(2).max(150),
  email: z.string().email(),
  password: z.string().min(8).max(72),
});

export function RegisterForm() {
  const t = useTranslations("auth");
  const router = useRouter();
  const register = useRegister();
  const login = useLogin();

  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", email: "", password: "" },
  });

  async function onSubmit(values: z.infer<typeof schema>) {
    try {
      await register.mutateAsync(values);
      await login.mutateAsync({ email: values.email, password: values.password });
      router.push("/");
    } catch (err) {
      const message = err instanceof ApiError ? err.message : "Registration failed";
      toast.error(message);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("name")}</FormLabel>
              <FormControl>
                <Input autoComplete="name" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
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
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("password")}</FormLabel>
              <FormControl>
                <Input type="password" autoComplete="new-password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit" className="w-full" disabled={register.isPending || login.isPending}>
          {t("submitRegister")}
        </Button>
      </form>
    </Form>
  );
}
