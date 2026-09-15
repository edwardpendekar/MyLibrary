import { getTranslations } from "next-intl/server";

import { SearchPageClient } from "@/features/search/components/search-page-client";

export default async function SearchPage({
  searchParams,
}: PageProps<"/[locale]/search">) {
  const t = await getTranslations("search");
  const { q } = await searchParams;
  const initialQuery = typeof q === "string" ? q : "";

  return (
    <div className="mx-auto max-w-5xl space-y-6 px-4 py-8">
      <h1 className="text-2xl font-bold">{t("title")}</h1>
      <SearchPageClient initialQuery={initialQuery} />
    </div>
  );
}
