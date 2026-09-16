import { getTranslations } from "next-intl/server";

import { ImportUploader } from "@/features/admin/components/import-uploader";
import { ImportHistory } from "@/features/admin/components/import-history";

export default async function AdminImportPage() {
  const t = await getTranslations("admin");

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold">{t("import")}</h1>
      <ImportUploader />
      <div className="space-y-3">
        <h2 className="text-lg font-semibold">{t("import_.history")}</h2>
        <ImportHistory />
      </div>
    </div>
  );
}
