import { getTranslations } from "next-intl/server";

import { SearchBar } from "@/features/search/components/search-bar";
import { PopularRail } from "@/features/books/components/popular-rail";
import { BookGrid } from "@/features/books/components/book-grid";

export default async function HomePage() {
  const t = await getTranslations("home");

  return (
    <div className="mx-auto max-w-7xl space-y-10 px-4 py-8">
      <section className="space-y-4 text-center">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">{t("title")}</h1>
        <p className="mx-auto max-w-2xl text-muted-foreground">{t("subtitle")}</p>
        <div className="flex justify-center">
          <SearchBar />
        </div>
      </section>

      <PopularRail />

      <section className="space-y-4">
        <h2 className="text-lg font-semibold">{t("allBooks")}</h2>
        <BookGrid />
      </section>
    </div>
  );
}
