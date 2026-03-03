"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { clearAdminToken, getAdminToken } from "@/lib/admin";

export default function AdminShell({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    const t = getAdminToken();
    if (!t) router.replace("/admin/login");
  }, [router]);

  return (
    <div className="space-y-6">
      <nav className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-slate-200 p-4">
        <div className="flex items-center gap-3">
          <Link className="font-semibold text-vexo-black hover:text-vexo-teal" href="/">
            Vexo
          </Link>
          <span className="text-slate-300">/</span>
          <span className="text-sm text-slate-700">Admin</span>
        </div>

        <div className="flex items-center gap-3 text-sm">
          <Link
            className={pathname.startsWith("/admin/products") ? "underline" : "hover:underline"}
            href="/admin/products"
          >
            Products
          </Link>
          <Link
            className={pathname.startsWith("/admin/orders") ? "underline" : "hover:underline"}
            href="/admin/orders"
          >
            Orders
          </Link>
          <button
            className="rounded-lg border border-slate-300 px-3 py-1 hover:bg-slate-50"
            onClick={() => {
              clearAdminToken();
              router.push("/admin/login");
            }}
            type="button"
          >
            Logout
          </button>
        </div>
      </nav>

      {children}
    </div>
  );
}

