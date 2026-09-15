"use client";

import { useState } from "react";

import { SearchBar } from "@/features/search/components/search-bar";
import { SearchResults } from "@/features/search/components/search-results";
import { useDebounce } from "@/hooks/use-debounce";

export function SearchPageClient({ initialQuery }: { initialQuery: string }) {
  const [query, setQuery] = useState(initialQuery);
  const debounced = useDebounce(query, 300);

  return (
    <div className="space-y-6">
      <SearchBar initialValue={initialQuery} onChange={setQuery} />
      <SearchResults query={debounced} />
    </div>
  );
}
