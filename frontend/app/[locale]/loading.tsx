import { Skeleton } from "@/components/ui/skeleton";

// Global fallback while a route segment's Server Component data is loading.
// Route-specific loading.tsx files (books/[slug], search, admin, etc.) take
// precedence over this one for their own segment.
export default function Loading() {
  return (
    <div className="mx-auto max-w-7xl space-y-6 px-4 py-8">
      <Skeleton className="h-8 w-64" />
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
        {Array.from({ length: 12 }).map((_, i) => (
          <Skeleton key={i} className="aspect-2/3 w-full rounded-md" />
        ))}
      </div>
    </div>
  );
}
