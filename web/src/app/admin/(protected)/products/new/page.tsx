"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { adminCreateProduct, getAdminToken } from "@/lib/admin";

export default function AdminNewProductPage() {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);

  const [slug, setSlug] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [status, setStatus] = useState("draft");

  const [sku, setSku] = useState("");
  const [title, setTitle] = useState("");
  const [size, setSize] = useState("");
  const [color, setColor] = useState("");
  const [priceINR, setPriceINR] = useState(0);
  const [onHand, setOnHand] = useState(0);

  return (
    <main className="space-y-4">
      <h1 className="text-2xl font-semibold">New product</h1>

      <form
        className="rounded-2xl border border-slate-200 p-4 space-y-4"
        onSubmit={(e) => {
          e.preventDefault();
          setError(null);
          const token = getAdminToken();
          if (!token) return;
          startTransition(async () => {
            try {
              const p = await adminCreateProduct(token, {
                slug,
                name,
                description,
                status,
                variants: [
                  {
                    sku,
                    title,
                    size,
                    color,
                    price_inr: priceINR,
                    on_hand: onHand
                  }
                ]
              });
              router.push(`/admin/products/${p.id}`);
            } catch (err) {
              setError(err instanceof Error ? err.message : "Create failed");
            }
          });
        }}
      >
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="space-y-1">
            <div className="text-sm font-medium">Slug</div>
            <input
              className="w-full rounded-lg border border-slate-300 px-3 py-2"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="classic-tee"
              required
            />
          </div>
          <div className="space-y-1">
            <div className="text-sm font-medium">Status</div>
            <select
              className="w-full rounded-lg border border-slate-300 px-3 py-2"
              value={status}
              onChange={(e) => setStatus(e.target.value)}
            >
              <option value="draft">draft</option>
              <option value="active">active</option>
              <option value="archived">archived</option>
            </select>
          </div>
          <div className="space-y-1 sm:col-span-2">
            <div className="text-sm font-medium">Name</div>
            <input
              className="w-full rounded-lg border border-slate-300 px-3 py-2"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-1 sm:col-span-2">
            <div className="text-sm font-medium">Description</div>
            <textarea
              className="w-full rounded-lg border border-slate-300 px-3 py-2"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>
        </div>

        <div className="border-t border-slate-100 pt-4 space-y-3">
          <div className="font-medium">First variant</div>
          <div className="grid gap-3 sm:grid-cols-2">
            <div className="space-y-1">
              <div className="text-sm font-medium">SKU</div>
              <input
                className="w-full rounded-lg border border-slate-300 px-3 py-2"
                value={sku}
                onChange={(e) => setSku(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1">
              <div className="text-sm font-medium">Title</div>
              <input
                className="w-full rounded-lg border border-slate-300 px-3 py-2"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Black / M"
              />
            </div>
            <div className="space-y-1">
              <div className="text-sm font-medium">Size</div>
              <input
                className="w-full rounded-lg border border-slate-300 px-3 py-2"
                value={size}
                onChange={(e) => setSize(e.target.value)}
                placeholder="M"
              />
            </div>
            <div className="space-y-1">
              <div className="text-sm font-medium">Color</div>
              <input
                className="w-full rounded-lg border border-slate-300 px-3 py-2"
                value={color}
                onChange={(e) => setColor(e.target.value)}
                placeholder="Black"
              />
            </div>
            <div className="space-y-1">
              <div className="text-sm font-medium">Price (paise)</div>
              <input
                className="w-full rounded-lg border border-slate-300 px-3 py-2"
                type="number"
                min={0}
                value={priceINR}
                onChange={(e) => setPriceINR(Number(e.target.value))}
                required
              />
            </div>
            <div className="space-y-1">
              <div className="text-sm font-medium">On hand</div>
              <input
                className="w-full rounded-lg border border-slate-300 px-3 py-2"
                type="number"
                min={0}
                value={onHand}
                onChange={(e) => setOnHand(Number(e.target.value))}
                required
              />
            </div>
          </div>
        </div>

        <button
          className="rounded-lg bg-slate-900 px-3 py-2 text-white disabled:opacity-50"
          disabled={isPending}
          type="submit"
        >
          {isPending ? "Creating…" : "Create product"}
        </button>

        {error ? (
          <div className="rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
            {error}
          </div>
        ) : null}
      </form>
    </main>
  );
}

