"use client";

import { useTranslations } from "next-intl";
import { BarChart3, BookOpen, FileClock, FolderTree, Languages, Upload, Users } from "lucide-react";

import { Link } from "@/i18n/navigation";
import { usePathname } from "@/i18n/navigation";
import { cn } from "@/lib/utils";

export function AdminSidebar() {
  const t = useTranslations("admin");
  const pathname = usePathname();

  const items = [
    { href: "/admin", label: t("dashboard"), icon: BarChart3 },
    { href: "/admin/books", label: t("books"), icon: BookOpen },
    { href: "/admin/import", label: t("import"), icon: Upload },
    { href: "/admin/categories", label: t("categories"), icon: FolderTree },
    { href: "/admin/languages", label: t("languages"), icon: Languages },
    { href: "/admin/users", label: t("users"), icon: Users },
    { href: "/admin/audit-log", label: t("auditLog"), icon: FileClock },
  ];

  return (
    <nav className="flex h-full flex-col gap-0.5 p-3">
      {items.map((item) => {
        const active = pathname === item.href;
        const Icon = item.icon;
        return (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              "flex items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-muted",
              active && "bg-muted font-medium"
            )}
          >
            <Icon className="size-4" />
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}
