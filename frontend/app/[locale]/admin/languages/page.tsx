"use client";

import { useState } from "react";
import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { useAdminLanguages, useCreateLanguage, useDeleteLanguage } from "@/hooks/admin/use-admin-reference";

export default function AdminLanguagesPage() {
  const t = useTranslations("admin.languagesPage");
  const tCommon = useTranslations("common");
  const { data: languages, isPending } = useAdminLanguages();
  const createLanguage = useCreateLanguage();
  const deleteLanguage = useDeleteLanguage();
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState({ code: "", name: "", native_name: "" });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t("title")}</h1>
        <Button onClick={() => setOpen(true)}>
          <Plus className="size-4" />
          {t("newLanguage")}
        </Button>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("colCode")}</TableHead>
              <TableHead>{t("colName")}</TableHead>
              <TableHead>{t("colNativeName")}</TableHead>
              <TableHead className="text-right">{tCommon("actions")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {!isPending &&
              languages?.map((l) => (
                <TableRow key={l.id}>
                  <TableCell>{l.code}</TableCell>
                  <TableCell>{l.name}</TableCell>
                  <TableCell>{l.native_name}</TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="icon-sm" onClick={() => deleteLanguage.mutate(l.id)}>
                      <Trash2 className="size-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
      </div>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("newLanguage")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label>{t("codeLabel")}</Label>
              <Input value={form.code} onChange={(e) => setForm({ ...form, code: e.target.value })} />
            </div>
            <div className="space-y-1.5">
              <Label>{t("nameLabel")}</Label>
              <Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
            </div>
            <div className="space-y-1.5">
              <Label>{t("nativeNameLabel")}</Label>
              <Input value={form.native_name} onChange={(e) => setForm({ ...form, native_name: e.target.value })} />
            </div>
          </div>
          <DialogFooter>
            <Button
              onClick={() =>
                createLanguage.mutate(form, {
                  onSuccess: () => {
                    setOpen(false);
                    setForm({ code: "", name: "", native_name: "" });
                  },
                })
              }
              disabled={!form.code || !form.name || createLanguage.isPending}
            >
              {tCommon("create")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
