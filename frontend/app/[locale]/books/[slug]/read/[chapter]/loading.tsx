import { Skeleton } from "@/components/ui/skeleton";

export default function Loading() {
  return (
    <div className="grid h-[calc(100vh-3.5rem)] grid-cols-1 md:grid-cols-[200px_1fr_280px]">
      <aside className="hidden border-r p-3 md:block">
        <Skeleton className="mb-3 h-5 w-20" />
        <div className="grid grid-cols-4 gap-1.5">
          {Array.from({ length: 12 }).map((_, i) => (
            <Skeleton key={i} className="h-9 w-full" />
          ))}
        </div>
      </aside>
      <main className="space-y-3 px-4 py-6 sm:px-8">
        <Skeleton className="h-8 w-48" />
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-5 w-full" />
        ))}
      </main>
      <aside className="hidden border-l lg:block" />
    </div>
  );
}
