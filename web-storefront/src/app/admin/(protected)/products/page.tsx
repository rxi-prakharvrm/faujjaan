"use client";

import Link from "next/link";
import { useEffect, useState, useTransition } from "react";
import type { AdminProduct } from "@/lib/admin";
import { adminListProducts, getAdminToken } from "@/lib/admin";
import { formatINR } from "@/lib/money";

export default function AdminProductsPage() {
  const [products, setProducts] = useState<AdminProduct[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    const token = getAdminToken();
    if (!token) return;
    startTransition(async () => {
      try {
        setError(null);
        const res = await adminListProducts(token);
        setProducts(res);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load products");
      }
    });
  }, []);

  return (
    <main className="space-y-4">
      <div className="flex items-center justify-between gap-4">
        <div className="space-y-1">
          <h1 className="text-2xl font-semibold">Products</h1>
          <p className="text-sm text-slate-600">{products.length} total</p>
        </div>
        <Link
          className="rounded-lg bg-slate-900 px-3 py-2 text-sm text-white"
          href="/admin/products/new"
        >
          New product
        </Link>
      </div>

      {error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
          {error}
        </div>
      ) : null}

      <div className="space-y-3">
        {products.map((p) => (
          <div key={p.id} className="rounded-2xl border border-slate-200 p-4">
            <div className="flex items-start justify-between gap-4">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <div className="font-medium">{p.name}</div>
                  <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-700">
                    {p.status}
                  </span>
                </div>
                <div className="text-sm text-slate-600">
                  <span className="font-mono">{p.slug}</span>
                </div>
              </div>
              <Link className="text-sm underline" href={`/admin/products/${p.id}`}>
                Edit
              </Link>
            </div>

            <div className="mt-3 overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="text-left text-slate-600">
                  <tr>
                    <th className="py-2 pr-3">SKU</th>
                    <th className="py-2 pr-3">Variant</th>
                    <th className="py-2 pr-3">Price</th>
                    <th className="py-2 pr-3">On hand</th>
                    <th className="py-2 pr-3">Reserved</th>
                  </tr>
                </thead>
                <tbody>
                  {p.variants.map((v) => (
                    <tr key={v.id} className="border-t border-slate-100">
                      <td className="py-2 pr-3 font-mono">{v.sku}</td>
                      <td className="py-2 pr-3">{v.title || `${v.color} / ${v.size}`}</td>
                      <td className="py-2 pr-3">{formatINR(v.price_inr)}</td>
                      <td className="py-2 pr-3">{v.on_hand}</td>
                      <td className="py-2 pr-3">{v.reserved}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))}
      </div>

      {isPending ? <div className="text-sm text-slate-500">Loading…</div> : null}
    </main>
  );
}

