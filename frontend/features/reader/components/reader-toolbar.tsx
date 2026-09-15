"use client";

import { useTranslations } from "next-intl";
import { Minus, Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useReaderPreferencesStore, type ReaderTranslation } from "@/store/reader-preferences-store";

export function ReaderToolbar() {
  const t = useTranslations("reader");
  const { fontSize, translation, increaseFontSize, decreaseFontSize, setTranslation } =
    useReaderPreferencesStore();

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 border-b bg-background/95 px-4 py-2">
      <div className="flex items-center gap-1">
        <span className="text-xs text-muted-foreground mr-1">{t("fontSize")}</span>
        <Button variant="outline" size="icon-sm" onClick={decreaseFontSize} aria-label="Decrease font size">
          <Minus className="size-3.5" />
        </Button>
        <span className="w-6 text-center text-xs tabular-nums">{fontSize}</span>
        <Button variant="outline" size="icon-sm" onClick={increaseFontSize} aria-label="Increase font size">
          <Plus className="size-3.5" />
        </Button>
      </div>

      <Tabs value={translation} onValueChange={(v) => setTranslation(v as ReaderTranslation)}>
        <TabsList>
          <TabsTrigger value="en">EN</TabsTrigger>
          <TabsTrigger value="id">ID</TabsTrigger>
          <TabsTrigger value="both">{t("compareTranslation")}</TabsTrigger>
        </TabsList>
      </Tabs>
    </div>
  );
}
