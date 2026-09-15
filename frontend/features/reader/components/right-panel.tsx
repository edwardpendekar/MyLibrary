"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Trash2 } from "lucide-react";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/store/auth-store";
import { useBookmarks, useDeleteBookmark } from "@/hooks/use-bookmarks";
import { useDeleteNote, useNotes } from "@/hooks/use-notes";
import { useSearchVerses } from "@/hooks/use-search";
import { useDebounce } from "@/hooks/use-debounce";
import { sanitizeSnippet } from "@/lib/sanitize-snippet";
import { Link } from "@/i18n/navigation";

export function RightPanel({ bookId, bookSlug }: { bookId: number; bookSlug: string }) {
  const t = useTranslations("reader");
  const user = useAuthStore((s) => s.user);

  return (
    <Tabs defaultValue="bookmarks" className="flex h-full flex-col">
      <TabsList className="m-2">
        <TabsTrigger value="bookmarks">{t("bookmarks")}</TabsTrigger>
        <TabsTrigger value="notes">{t("notes")}</TabsTrigger>
        <TabsTrigger value="search">{t("search")}</TabsTrigger>
      </TabsList>

      <TabsContent value="bookmarks" className="flex-1 overflow-hidden">
        {user ? <BookmarksTab bookId={bookId} bookSlug={bookSlug} /> : <SignInPrompt />}
      </TabsContent>
      <TabsContent value="notes" className="flex-1 overflow-hidden">
        {user ? <NotesTab bookId={bookId} /> : <SignInPrompt />}
      </TabsContent>
      <TabsContent value="search" className="flex-1 overflow-hidden">
        <InBookSearchTab bookId={bookId} bookSlug={bookSlug} />
      </TabsContent>
    </Tabs>
  );
}

function SignInPrompt() {
  const t = useTranslations("auth");
  return (
    <p className="p-4 text-sm text-muted-foreground">
      <Link href="/login" className="underline">
        {t("loginTitle")}
      </Link>
    </p>
  );
}

function BookmarksTab({ bookId, bookSlug }: { bookId: number; bookSlug: string }) {
  const t = useTranslations("reader");
  const { data: bookmarks } = useBookmarks(bookId);
  const deleteBookmark = useDeleteBookmark(bookId);

  if (!bookmarks || bookmarks.length === 0) {
    return <p className="p-4 text-sm text-muted-foreground">{t("emptyBookmarks")}</p>;
  }

  return (
    <ScrollArea className="h-full">
      <ul className="space-y-1 p-2">
        {bookmarks.map((b) => (
          <li key={b.id} className="flex items-center justify-between rounded-md px-2 py-1.5 text-sm hover:bg-muted">
            <Link href={`/books/${bookSlug}/read/${b.chapter_id ?? 1}#v${b.verse_id ?? ""}`}>
              {b.label || `Bookmark #${b.id}`}
            </Link>
            <Button variant="ghost" size="icon-xs" onClick={() => deleteBookmark.mutate(b.id)}>
              <Trash2 className="size-3.5" />
            </Button>
          </li>
        ))}
      </ul>
    </ScrollArea>
  );
}

function NotesTab({ bookId }: { bookId: number }) {
  const t = useTranslations("reader");
  const { data: notes } = useNotes(bookId);
  const deleteNote = useDeleteNote(bookId);

  if (!notes || notes.length === 0) {
    return <p className="p-4 text-sm text-muted-foreground">{t("emptyNotes")}</p>;
  }

  return (
    <ScrollArea className="h-full">
      <ul className="space-y-2 p-2">
        {notes.map((n) => (
          <li key={n.id} className="rounded-md border p-2 text-sm">
            <div className="flex items-start justify-between gap-2">
              <p className="whitespace-pre-wrap">{n.content}</p>
              <Button variant="ghost" size="icon-xs" onClick={() => deleteNote.mutate(n.id)}>
                <Trash2 className="size-3.5" />
              </Button>
            </div>
          </li>
        ))}
      </ul>
    </ScrollArea>
  );
}

function InBookSearchTab({ bookId, bookSlug }: { bookId: number; bookSlug: string }) {
  const t = useTranslations("reader");
  const [query, setQuery] = useState("");
  const debounced = useDebounce(query);
  const { data: hits, isPending } = useSearchVerses(debounced, bookId);

  return (
    <div className="flex h-full flex-col">
      <div className="p-2">
        <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder={t("search")} />
      </div>
      <ScrollArea className="flex-1">
        <ul className="space-y-1 p-2">
          {!isPending &&
            hits?.map((hit) => (
              <li key={hit.verse_id}>
                <Link
                  href={`/books/${bookSlug}/read/${hit.chapter_number}#v${hit.verse_number}`}
                  className="block rounded-md p-2 text-sm hover:bg-muted"
                >
                  <span className="font-medium">
                    {hit.chapter_number}:{hit.verse_number}
                  </span>{" "}
                  <span
                    className="text-muted-foreground"
                    dangerouslySetInnerHTML={{ __html: sanitizeSnippet(hit.snippet) }}
                  />
                </Link>
              </li>
            ))}
        </ul>
      </ScrollArea>
    </div>
  );
}
