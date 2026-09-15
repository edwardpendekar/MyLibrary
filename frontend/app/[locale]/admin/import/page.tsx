import { ImportUploader } from "@/features/admin/components/import-uploader";
import { ImportHistory } from "@/features/admin/components/import-history";

export default function AdminImportPage() {
  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold">Import</h1>
      <ImportUploader />
      <div className="space-y-3">
        <h2 className="text-lg font-semibold">History</h2>
        <ImportHistory />
      </div>
    </div>
  );
}
