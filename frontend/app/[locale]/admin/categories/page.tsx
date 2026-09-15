"use client";

import { useState } from "react";
import { Plus, Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { useAdminCategories, useCreateCategory, useDeleteCategory } from "@/hooks/admin/use-admin-reference";

export default function AdminCategoriesPage() {
  const { data: categories, isPending } = useAdminCategories();
  const createCategory = useCreateCategory();
  const deleteCategory = useDeleteCategory();
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState({ name_en: "", name_id: "" });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Categories</h1>
        <Button onClick={() => setOpen(true)}>
          <Plus className="size-4" />
          New category
        </Button>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Slug</TableHead>
              <TableHead>Name (EN)</TableHead>
              <TableHead>Name (ID)</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {!isPending &&
              categories?.map((c) => (
                <TableRow key={c.id}>
                  <TableCell>{c.slug}</TableCell>
                  <TableCell>{c.name_en}</TableCell>
                  <TableCell>{c.name_id}</TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="icon-sm" onClick={() => deleteCategory.mutate(c.id)}>
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
            <DialogTitle>New category</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label>Name (English)</Label>
              <Input value={form.name_en} onChange={(e) => setForm({ ...form, name_en: e.target.value })} />
            </div>
            <div className="space-y-1.5">
              <Label>Name (Indonesian)</Label>
              <Input value={form.name_id} onChange={(e) => setForm({ ...form, name_id: e.target.value })} />
            </div>
          </div>
          <DialogFooter>
            <Button
              onClick={() =>
                createCategory.mutate(form, {
                  onSuccess: () => {
                    setOpen(false);
                    setForm({ name_en: "", name_id: "" });
                  },
                })
              }
              disabled={!form.name_en || !form.name_id || createCategory.isPending}
            >
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
