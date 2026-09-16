"use client";

import { useRef } from "react";
import Image from "next/image";
import { FileText, Upload } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export function FileUploadCard({
  title,
  accept,
  currentUrl,
  isImage,
  isUploading,
  onSelect,
}: {
  title: string;
  accept: string;
  currentUrl?: string | null;
  isImage?: boolean;
  isUploading?: boolean;
  onSelect: (file: File) => void;
}) {
  const t = useTranslations("admin.fileUpload");
  const inputRef = useRef<HTMLInputElement | null>(null);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm">{title}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {currentUrl && isImage && (
          <div className="relative h-40 w-28 overflow-hidden rounded-md border">
            <Image src={currentUrl} alt={title} fill className="object-cover" />
          </div>
        )}
        {currentUrl && !isImage && (
          <a href={currentUrl} target="_blank" rel="noreferrer" className="flex items-center gap-2 text-sm text-primary hover:underline">
            <FileText className="size-4" />
            {t("viewCurrentFile")}
          </a>
        )}
        <input
          ref={inputRef}
          type="file"
          accept={accept}
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0];
            if (file) onSelect(file);
            e.target.value = "";
          }}
        />
        <Button type="button" variant="outline" size="sm" disabled={isUploading} onClick={() => inputRef.current?.click()}>
          <Upload className="size-4" />
          {isUploading ? t("uploading") : t("upload")}
        </Button>
      </CardContent>
    </Card>
  );
}
