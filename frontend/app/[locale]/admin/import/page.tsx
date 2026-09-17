import { getTranslations } from "next-intl/server";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ImportUploader } from "@/features/admin/components/import-uploader";
import { TranslateUploader } from "@/features/admin/components/translate-uploader";
import { ImportHistory } from "@/features/admin/components/import-history";

export default async function AdminImportPage() {
  const t = await getTranslations("admin");
  const tImport = await getTranslations("admin.import_");

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold">{t("import")}</h1>

      <Tabs defaultValue="spreadsheet">
        <TabsList>
          <TabsTrigger value="spreadsheet">{tImport("tabSpreadsheet")}</TabsTrigger>
          <TabsTrigger value="translate">{tImport("tabTranslate")}</TabsTrigger>
        </TabsList>
        <TabsContent value="spreadsheet">
          <ImportUploader />
        </TabsContent>
        <TabsContent value="translate">
          <TranslateUploader />
        </TabsContent>
      </Tabs>

      <div className="space-y-3">
        <h2 className="text-lg font-semibold">{tImport("history")}</h2>
        <ImportHistory />
      </div>
    </div>
  );
}
