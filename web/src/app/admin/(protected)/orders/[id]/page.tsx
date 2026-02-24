"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState, useTransition } from "react";
import { adminGetOrder, getAdminToken } from "@/lib/admin";
import { formatINR } from "@/lib/money";

type OrderDetail = {
  id: string;
  status: string;
  subtotal_inr: number;
  shipping_inr: number;
  tax_inr: number;
  total_inr: number;
  customer_name: string;
  customer_phone: string;
  customer_email: string;
  shipping_address: Record<string, unknown>;
  payment_status: string;
  razorpay_order_id: string;
  items: Array<{
    sku: string;
    product_name: string;
    variant_title: string;
    unit_price_inr: number;
    quantity: number;
    line_total_inr: number;
  }>;
  created_at: string;
};

export default function AdminOrderDetailPage() {
  const params = useParams<{ id: string }>();
  const orderId = params.id;
  const [order, setOrder] = useState<OrderDetail | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    const token = getAdminToken();
    if (!token) return;
    startTransition(async () => {
      try {
        setError(null);
        const o = (await adminGetOrder(token, orderId)) as OrderDetail;
        setOrder(o);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load order");
      }
    });
  }, [orderId]);

  return (
    <main className="space-y-4">
      <Link className="text-sm underline" href="/admin/orders">
        ← Orders
      </Link>

      {error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
          {error}
        </div>
      ) : null}

      {!order ? (
        <div className="text-slate-600">Loading…</div>
      ) : (
        <>
          <div className="rounded-2xl border border-slate-200 p-4 space-y-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div className="font-medium">Order {order.id.slice(0, 8)}…</div>
              <div className="text-sm text-slate-700">{order.status}</div>
            </div>
            <div className="text-sm text-slate-600">
              Created {new Date(order.created_at).toLocaleString()}
            </div>
            <div className="grid gap-2 sm:grid-cols-2 text-sm">
              <div>
                <div className="text-slate-600">Customer</div>
                <div className="font-medium">{order.customer_name}</div>
                <div>{order.customer_phone}</div>
                {order.customer_email ? <div>{order.customer_email}</div> : null}
              </div>
              <div>
                <div className="text-slate-600">Payment</div>
                <div className="font-medium">{order.payment_status}</div>
                {order.razorpay_order_id ? (
                  <div className="font-mono text-xs">{order.razorpay_order_id}</div>
                ) : (
                  <div className="text-xs text-slate-500">Razorpay order not created yet</div>
                )}
              </div>
            </div>
          </div>

          <div className="rounded-2xl border border-slate-200 overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-left text-slate-600">
                <tr>
                  <th className="px-4 py-3">SKU</th>
                  <th className="px-4 py-3">Item</th>
                  <th className="px-4 py-3">Qty</th>
                  <th className="px-4 py-3">Line</th>
                </tr>
              </thead>
              <tbody>
                {order.items.map((it, idx) => (
                  <tr key={idx} className="border-t border-slate-100">
                    <td className="px-4 py-3 font-mono">{it.sku}</td>
                    <td className="px-4 py-3">
                      <div className="font-medium">{it.product_name}</div>
                      <div className="text-slate-600">{it.variant_title}</div>
                      <div className="text-slate-600">{formatINR(it.unit_price_inr)}</div>
                    </td>
                    <td className="px-4 py-3">{it.quantity}</td>
                    <td className="px-4 py-3">{formatINR(it.line_total_inr)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="rounded-2xl border border-slate-200 p-4 space-y-1 text-sm">
            <div className="flex items-center justify-between">
              <div className="text-slate-600">Subtotal</div>
              <div className="font-medium">{formatINR(order.subtotal_inr)}</div>
            </div>
            <div className="flex items-center justify-between">
              <div className="text-slate-600">Shipping</div>
              <div className="font-medium">{formatINR(order.shipping_inr)}</div>
            </div>
            <div className="flex items-center justify-between">
              <div className="text-slate-600">Tax</div>
              <div className="font-medium">{formatINR(order.tax_inr)}</div>
            </div>
            <div className="flex items-center justify-between border-t border-slate-100 pt-2">
              <div className="text-slate-600">Total</div>
              <div className="font-semibold">{formatINR(order.total_inr)}</div>
            </div>
          </div>

          {isPending ? <div className="text-sm text-slate-500">Loading…</div> : null}
        </>
      )}
    </main>
  );
}

