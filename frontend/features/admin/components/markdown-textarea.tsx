"use client";

import { useRef } from "react";
import { useTranslations } from "next-intl";
import { Bold, Italic } from "lucide-react";
import ReactMarkdown from "react-markdown";

import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { wrapMarkdownSelection } from "@/lib/wrap-markdown-selection";

/**
 * A plain <textarea> that edits Markdown, with a Bold/Italic toolbar (wraps
 * the current selection) and a live Preview tab — deliberately not a full
 * WYSIWYG editor: the stored value is just Markdown text, which is what
 * ReactMarkdown renders wherever a description is displayed, and Markdown
 * doesn't carry the raw-HTML injection risk a real HTML editor would.
 */
export function MarkdownTextarea({
  value,
  onChange,
  rows = 4,
}: {
  value: string;
  onChange: (value: string) => void;
  rows?: number;
}) {
  const t = useTranslations("admin.markdown");
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);

  function wrapSelection(marker: string) {
    const el = textareaRef.current;
    if (!el) return;
    const result = wrapMarkdownSelection(value, el.selectionStart, el.selectionEnd, marker);
    onChange(result.next);
    // Restore focus + selection (inside the markers) after React re-renders.
    requestAnimationFrame(() => {
      el.focus();
      el.setSelectionRange(result.selectionStart, result.selectionEnd);
    });
  }

  return (
    <Tabs defaultValue="write">
      <div className="flex items-center justify-between">
        <div className="flex gap-0.5">
          <Button
            type="button"
            variant="ghost"
            size="icon-xs"
            aria-label={t("bold")}
            onClick={() => wrapSelection("**")}
          >
            <Bold className="size-3.5" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon-xs"
            aria-label={t("italic")}
            onClick={() => wrapSelection("*")}
          >
            <Italic className="size-3.5" />
          </Button>
        </div>
        <TabsList>
          <TabsTrigger value="write">{t("write")}</TabsTrigger>
          <TabsTrigger value="preview">{t("preview")}</TabsTrigger>
        </TabsList>
      </div>

      <TabsContent value="write">
        <Textarea
          ref={textareaRef}
          rows={rows}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      </TabsContent>
      <TabsContent value="preview">
        <div
          className="prose prose-sm dark:prose-invert min-h-24 max-w-none rounded-md border px-3 py-2 prose-p:leading-relaxed"
          style={{ minHeight: `${rows * 1.5}rem` }}
        >
          {value.trim() ? (
            <ReactMarkdown>{value}</ReactMarkdown>
          ) : (
            <p className="text-sm text-muted-foreground">{t("nothingToPreview")}</p>
          )}
        </div>
      </TabsContent>
    </Tabs>
  );
}
