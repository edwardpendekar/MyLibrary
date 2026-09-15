"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { useCreateNote } from "@/hooks/use-notes";

export function AddNoteDialog({
  bookId,
  verseId,
  onOpenChange,
}: {
  bookId: number;
  verseId: number | null;
  onOpenChange: (open: boolean) => void;
}) {
  const t = useTranslations("reader");
  const [content, setContent] = useState("");
  const createNote = useCreateNote(bookId);

  return (
    <Dialog open={verseId !== null} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("addNote")}</DialogTitle>
        </DialogHeader>
        <Textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder={t("notePlaceholder")}
          rows={5}
        />
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            {t("cancel")}
          </Button>
          <Button
            disabled={!content.trim() || createNote.isPending}
            onClick={() => {
              if (!verseId) return;
              createNote.mutate(
                { verse_id: verseId, content: content.trim() },
                {
                  onSuccess: () => {
                    toast.success(t("save"));
                    setContent("");
                    onOpenChange(false);
                  },
                }
              );
            }}
          >
            {t("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
