"use client";

import { useMemo, useState, useTransition } from "react";
import type { Product, Variant } from "@/lib/api";
import { createCart, upsertCartItem } from "@/lib/api";
import { getStoredCartId, setStoredCartId } from "@/lib/cart";
import { formatINR } from "@/lib/money";

export default function ProductClient({ product }: { product: Product }) {
  const variants = product.variants;
  const [variantId, setVariantId] = useState<string>(variants[0]?.id ?? "");
  const [qty, setQty] = useState(1);
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);
  const selected = useMemo<Variant | undefined>(
    () => variants.find((v) => v.id === variantId),
    [variants, variantId]
  );

  async function addToCart() {
    setError(null);
    if (!selected) return;
    if (qty <= 0) return;

    startTransition(async () => {
      try {
        let cartId = getStoredCartId();
        if (!cartId) {
          cartId = await createCart();
          setStoredCartId(cartId);
        }
        await upsertCartItem(cartId, selected.id, qty);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to add to cart");
      }
    });
  }

  const available =
    selected ? Math.max(0, selected.on_hand - selected.reserved) : 0;

  return (
    <section className="rounded-2xl border border-vexo-gray/50 bg-white p-6 space-y-6">
      <div className="space-y-4">
        <div>
          <label className="text-sm font-medium text-vexo-black">Variant</label>
          <select
            className="mt-2 w-full rounded-lg border border-vexo-gray/80 bg-white px-4 py-3 text-vexo-black transition-colors focus:border-vexo-teal focus:outline-none focus:ring-2 focus:ring-vexo-teal/20"
            value={variantId}
            onChange={(e) => setVariantId(e.target.value)}
          >
            {variants.map((v) => (
              <option key={v.id} value={v.id}>
                {v.title || `${v.color} / ${v.size}`} — {formatINR(v.price_inr)}
              </option>
            ))}
          </select>
          <p className="mt-2 text-sm text-vexo-brown">
            {available > 0 ? `${available} available` : "Out of stock"}
          </p>
        </div>

        <div>
          <label className="text-sm font-medium text-vexo-black">Quantity</label>
          <input
            className="mt-2 w-full rounded-lg border border-vexo-gray/80 bg-white px-4 py-3 text-vexo-black transition-colors focus:border-vexo-teal focus:outline-none focus:ring-2 focus:ring-vexo-teal/20"
            type="number"
            min={1}
            max={20}
            value={qty}
            onChange={(e) => setQty(Number(e.target.value))}
          />
        </div>

        <button
          className="w-full rounded-lg bg-vexo-black px-4 py-3.5 font-medium text-white transition-all hover:bg-vexo-brown disabled:opacity-50 disabled:cursor-not-allowed"
          disabled={isPending || !selected || available <= 0}
          onClick={addToCart}
        >
          {isPending ? "Adding..." : "Add to cart"}
        </button>
      </div>

      {error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
          {error}
        </div>
      ) : null}
    </section>
  );
}
