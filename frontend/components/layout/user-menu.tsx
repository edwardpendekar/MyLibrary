"use client";

import { useTranslations } from "next-intl";
import { LogOut, User as UserIcon } from "lucide-react";

import { Link } from "@/i18n/navigation";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useAuthStore } from "@/store/auth-store";
import { useLogout } from "@/hooks/use-auth";

export function UserMenu() {
  const t = useTranslations("nav");
  const user = useAuthStore((s) => s.user);
  const isLoading = useAuthStore((s) => s.isLoading);
  const logout = useLogout();

  if (isLoading) return <div className="size-8" />;

  if (!user) {
    // Real navigation — styled via buttonVariants on the Link itself so it
    // keeps link semantics/keyboard behavior (see SiteHeader for why).
    return (
      <div className="flex items-center gap-1">
        <Link href="/login" className={buttonVariants({ variant: "ghost", size: "sm" })}>
          {t("login")}
        </Link>
        <Link href="/register" className={buttonVariants({ size: "sm" })}>
          {t("register")}
        </Link>
      </div>
    );
  }

  const initials = user.name
    .split(" ")
    .map((p) => p[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button variant="ghost" size="icon" className="rounded-full" aria-label="User menu">
            <Avatar className="size-8">
              <AvatarFallback>{initials || <UserIcon className="size-4" />}</AvatarFallback>
            </Avatar>
          </Button>
        }
      />
      <DropdownMenuContent align="end">
        <DropdownMenuGroup>
          <DropdownMenuLabel className="flex flex-col">
            <span className="font-medium">{user.name}</span>
            <span className="text-muted-foreground text-xs">{user.email}</span>
          </DropdownMenuLabel>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => logout.mutate()}>
          <LogOut className="size-4" />
          {t("logout")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
