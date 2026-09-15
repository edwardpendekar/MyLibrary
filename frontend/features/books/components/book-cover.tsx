import Image from "next/image";
import { BookOpen } from "lucide-react";

import { COVER_BLUR_DATA_URL } from "@/lib/image-placeholder";
import { cn } from "@/lib/utils";

export function BookCover({
  src,
  alt,
  className,
  sizes = "(min-width: 1024px) 200px, (min-width: 640px) 160px, 45vw",
  priority = false,
}: {
  src?: string | null;
  alt: string;
  className?: string;
  sizes?: string;
  priority?: boolean;
}) {
  return (
    <div
      className={cn(
        "relative aspect-2/3 w-full overflow-hidden rounded-md bg-muted",
        className
      )}
    >
      {src ? (
        <Image
          src={src}
          alt={alt}
          fill
          sizes={sizes}
          priority={priority}
          placeholder="blur"
          blurDataURL={COVER_BLUR_DATA_URL}
          className="object-cover transition-transform duration-300 group-hover:scale-105"
        />
      ) : (
        <div className="flex h-full w-full items-center justify-center text-muted-foreground">
          <BookOpen className="size-8" />
        </div>
      )}
    </div>
  );
}
