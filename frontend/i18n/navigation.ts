import { createNavigation } from "next-intl/navigation";
import { routing } from "./routing";

// Locale-aware Link/router/redirect/pathnames, so feature code never has to
// manually prefix hrefs with "/en" or "/id".
export const { Link, redirect, usePathname, useRouter, getPathname } =
  createNavigation(routing);
