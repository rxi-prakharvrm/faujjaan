"use client";

import Link from "next/link";
import { useEffect, useState, useTransition } from "react";
import type { AdminOrderSummary } from "@/lib/admin";
import { adminListOrders, getAdminToken } from "@/lib/admin";
import { formatINR } from "@/lib/money";

export default function AdminOrdersPage() {
  const [orders, setOrders] = useState<AdminOrderSummary[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    const token = getAdminToken();
    if (!token) return;
    startTransition(async () => {
      try {
        setError(null);
        const res = await adminListOrders(token);
        setOrders(res);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load orders");
      }
    });
  }, []);

  return (
    <main className="space-y-4">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold">Orders</h1>
        <p className="text-sm text-slate-600">{orders.length} recent</p>
      </div>

      {error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
          {error}
        </div>
      ) : null}

      <div className="rounded-2xl border border-slate-200 overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-left text-slate-600">
            <tr>
              <th className="px-4 py-3">Order</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Total</th>
              <th className="px-4 py-3">Created</th>
            </tr>
          </thead>
          <tbody>
            {orders.map((o) => (
              <tr key={o.id} className="border-t border-slate-100">
                <td className="px-4 py-3">
                  <Link className="underline" href={`/admin/orders/${o.id}`}>
                    {o.id.slice(0, 8)}…
                  </Link>
                </td>
                <td className="px-4 py-3">{o.status}</td>
                <td className="px-4 py-3">{formatINR(o.total_inr)}</td>
                <td className="px-4 py-3">
                  {new Date(o.created_at).toLocaleString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {isPending ? <div className="text-sm text-slate-500">Loading…</div> : null}
    </main>
  );
}

