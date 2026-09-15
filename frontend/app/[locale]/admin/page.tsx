"use client";

import { useTranslations } from "next-intl";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useAdminStats } from "@/hooks/admin/use-admin-stats";

export default function AdminDashboardPage() {
  const t = useTranslations("admin");
  const { data: stats, isPending } = useAdminStats();

  const cards = [
    { key: "totalBooks", value: stats?.total_books },
    { key: "publishedBooks", value: stats?.published_books },
    { key: "totalChapters", value: stats?.total_chapters },
    { key: "totalVerses", value: stats?.total_verses },
    { key: "totalUsers", value: stats?.total_users },
    { key: "importsLast30Day", value: stats?.imports_last_30_day },
  ] as const;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">{t("dashboard")}</h1>
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
        {cards.map((card) => (
          <Card key={card.key}>
            <CardHeader className="pb-2">
              <CardTitle className="text-xs font-normal text-muted-foreground">
                {t(`stats.${card.key}`)}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {isPending ? <Skeleton className="h-7 w-16" /> : <p className="text-2xl font-bold">{card.value ?? 0}</p>}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
