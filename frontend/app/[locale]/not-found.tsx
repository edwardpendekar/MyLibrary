import { getTranslations } from "next-intl/server";
import { BookX } from "lucide-react";

import { Link } from "@/i18n/navigation";
import { buttonVariants } from "@/components/ui/button";

export default async function NotFound() {
  const t = await getTranslations("book");
  const tCommon = await getTranslations("common");

  return (
    <div className="mx-auto flex max-w-md flex-col items-center gap-4 px-4 py-24 text-center">
      <BookX className="size-10 text-muted-foreground" />
      <h1 className="text-xl font-semibold">{t("notFound")}</h1>
      <Link href="/" className={buttonVariants()}>
        {tCommon("back")}
      </Link>
    </div>
  );
}
