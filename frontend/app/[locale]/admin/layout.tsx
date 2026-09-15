import { redirect } from "next/navigation";

import { apiServer } from "@/lib/api-server";
import { AdminSidebar } from "@/components/layout/admin-sidebar";
import type { User } from "@/types/api";

export default async function AdminLayout({ children }: LayoutProps<"/[locale]/admin">) {
  let user: User;
  try {
    const res = await apiServer.get<User>("/api/v1/me");
    user = res.data;
  } catch {
    redirect("/login");
  }

  if (user.role !== "admin" && user.role !== "editor") {
    redirect("/");
  }

  return (
    <div className="grid min-h-[calc(100vh-3.5rem)] grid-cols-1 md:grid-cols-[220px_1fr]">
      <aside className="hidden border-r md:block">
        <AdminSidebar />
      </aside>
      <div className="p-6">{children}</div>
    </div>
  );
}
