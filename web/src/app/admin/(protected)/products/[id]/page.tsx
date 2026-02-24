"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { type TransitionStartFunction, useEffect, useMemo, useState, useTransition } from "react";
import type { AdminProduct } from "@/lib/admin";
import {
  adminAdjustInventory,
  adminCreateVariant,
  adminListProducts,
  adminUpdateProduct,
  adminUpdateVariant,
  getAdminToken
} from "@/lib/admin";
import { formatINR } from "@/lib/money";

export default function AdminEditProductPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const productId = params.id;

  const [product, setProduct] = useState<AdminProduct | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  const token = useMemo(() => getAdminToken(), []);

  useEffect(() => {
    if (!token) return;
    startTransition(async () => {
      try {
        setError(null);
        const products = await adminListProducts(token);
        const p = products.find((x) => x.id === productId) ?? null;
        setProduct(p);
        if (!p) setError("Product not found");
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load product");
      }
    });
  }, [productId, token]);

  if (!token) return null;

  if (!product) {
    return (
      <main className="space-y-3">
        <div className="text-slate-600">Loading…</div>
        {error ? (
          <div className="rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
            {error}
          </div>
        ) : null}
      </main>
    );
  }

  return (
    <main className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <Link className="text-sm underline" href="/admin/products">
          ← Products
        </Link>
        <button
          className="text-sm underline text-slate-700"
          onClick={() => router.refresh()}
          type="button"
        >
          Refresh
        </button>
      </div>

      <section className="rounded-2xl border border-slate-200 p-4 space-y-4">
        <div className="font-medium">Product</div>
        <ProductEditor
          token={token}
          product={product}
          onSaved={(p) => setProduct(p)}
          setError={setError}
          isPending={isPending}
          startTransition={startTransition}
        />
      </section>

      <section className="rounded-2xl border border-slate-200 p-4 space-y-4">
        <div className="flex items-center justify-between">
          <div className="font-medium">Variants</div>
          <span className="text-sm text-slate-600">{product.variants.length}</span>
        </div>

        <div className="space-y-3">
          {product.variants.map((v) => (
            <VariantCard
              key={v.id}
              token={token}
              variant={v}
              onChanged={async () => {
                const products = await adminListProducts(token);
                const p = products.find((x) => x.id === productId) ?? null;
                if (p) setProduct(p);
              }}
              setError={setError}
            />
          ))}
        </div>
      </section>

      <section className="rounded-2xl border border-slate-200 p-4 space-y-4">
        <div className="font-medium">Add variant</div>
        <AddVariant
          token={token}
          productId={productId}
          setError={setError}
          onCreated={async () => {
            const products = await adminListProducts(token);
            const p = products.find((x) => x.id === productId) ?? null;
            if (p) setProduct(p);
          }}
        />
      </section>

      {error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
          {error}
        </div>
      ) : null}
    </main>
  );
}

function ProductEditor(props: {
  token: string;
  product: AdminProduct;
  onSaved: (p: AdminProduct) => void;
  setError: (s: string | null) => void;
  isPending: boolean;
  startTransition: TransitionStartFunction;
}) {
  const { token, product, onSaved, setError, isPending, startTransition } = props;

  const [slug, setSlug] = useState(product.slug);
  const [name, setName] = useState(product.name);
  const [description, setDescription] = useState(product.description);
  const [status, setStatus] = useState(product.status);

  return (
    <form
      className="grid gap-3 sm:grid-cols-2"
      onSubmit={(e) => {
        e.preventDefault();
        startTransition(async () => {
          try {
            setError(null);
            await adminUpdateProduct(token, product.id, {
              slug,
              name,
              description,
              status
            });
            onSaved({ ...product, slug, name, description, status });
          } catch (e) {
            setError(e instanceof Error ? e.message : "Update failed");
          }
        });
      }}
    >
      <label className="space-y-1">
        <div className="text-sm font-medium">Slug</div>
        <input
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          value={slug}
          onChange={(e) => setSlug(e.target.value)}
          required
        />
      </label>
      <label className="space-y-1">
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
      </label>
      <label className="space-y-1 sm:col-span-2">
        <div className="text-sm font-medium">Name</div>
        <input
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
      </label>
      <label className="space-y-1 sm:col-span-2">
        <div className="text-sm font-medium">Description</div>
        <textarea
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
        />
      </label>
      <div className="sm:col-span-2">
        <button
          className="rounded-lg bg-slate-900 px-3 py-2 text-sm text-white disabled:opacity-50"
          disabled={isPending}
          type="submit"
        >
          {isPending ? "Saving…" : "Save product"}
        </button>
      </div>
    </form>
  );
}

function VariantCard(props: {
  token: string;
  variant: AdminProduct["variants"][number];
  onChanged: () => Promise<void>;
  setError: (s: string | null) => void;
}) {
  const { token, variant, onChanged, setError } = props;
  const [isPending, startTransition] = useTransition();

  const [title, setTitle] = useState(variant.title);
  const [size, setSize] = useState(variant.size);
  const [color, setColor] = useState(variant.color);
  const [priceINR, setPriceINR] = useState(variant.price_inr);
  const [delta, setDelta] = useState(1);

  return (
    <div className="rounded-xl border border-slate-200 p-4 space-y-3">
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-1">
          <div className="font-mono text-sm">{variant.sku}</div>
          <div className="text-sm text-slate-600">
            {variant.title || `${variant.color} / ${variant.size}`}
          </div>
        </div>
        <div className="text-sm font-medium">{formatINR(variant.price_inr)}</div>
      </div>

      <form
        className="grid gap-3 sm:grid-cols-2"
        onSubmit={(e) => {
          e.preventDefault();
          startTransition(async () => {
            try {
              setError(null);
              await adminUpdateVariant(token, variant.id, {
                title,
                size,
                color,
                price_inr: priceINR
              });
              await onChanged();
            } catch (e) {
              setError(e instanceof Error ? e.message : "Update variant failed");
            }
          });
        }}
      >
        <label className="space-y-1">
          <div className="text-sm font-medium">Title</div>
          <input
            className="w-full rounded-lg border border-slate-300 px-3 py-2"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
        </label>
        <label className="space-y-1">
          <div className="text-sm font-medium">Price (paise)</div>
          <input
            className="w-full rounded-lg border border-slate-300 px-3 py-2"
            type="number"
            min={0}
            value={priceINR}
            onChange={(e) => setPriceINR(Number(e.target.value))}
          />
        </label>
        <label className="space-y-1">
          <div className="text-sm font-medium">Size</div>
          <input
            className="w-full rounded-lg border border-slate-300 px-3 py-2"
            value={size}
            onChange={(e) => setSize(e.target.value)}
          />
        </label>
        <label className="space-y-1">
          <div className="text-sm font-medium">Color</div>
          <input
            className="w-full rounded-lg border border-slate-300 px-3 py-2"
            value={color}
            onChange={(e) => setColor(e.target.value)}
          />
        </label>
        <div className="sm:col-span-2">
          <button
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50 disabled:opacity-50"
            disabled={isPending}
            type="submit"
          >
            {isPending ? "Saving…" : "Save variant"}
          </button>
        </div>
      </form>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 pt-3">
        <div className="text-sm">
          <span className="text-slate-600">On hand:</span> {variant.on_hand}{" "}
          <span className="text-slate-600">Reserved:</span> {variant.reserved}
        </div>
        <div className="flex items-center gap-2">
          <input
            className="w-20 rounded-lg border border-slate-300 px-2 py-1 text-sm"
            type="number"
            value={delta}
            onChange={(e) => setDelta(Number(e.target.value))}
          />
          <button
            className="rounded-lg bg-slate-900 px-3 py-2 text-sm text-white disabled:opacity-50"
            disabled={isPending || delta === 0}
            onClick={() => {
              startTransition(async () => {
                try {
                  setError(null);
                  await adminAdjustInventory(token, { variant_id: variant.id, delta });
                  await onChanged();
                } catch (e) {
                  setError(e instanceof Error ? e.message : "Inventory adjust failed");
                }
              });
            }}
            type="button"
          >
            Adjust inventory
          </button>
        </div>
      </div>
    </div>
  );
}

function AddVariant(props: {
  token: string;
  productId: string;
  setError: (s: string | null) => void;
  onCreated: () => Promise<void>;
}) {
  const { token, productId, setError, onCreated } = props;
  const [isPending, startTransition] = useTransition();

  const [sku, setSku] = useState("");
  const [title, setTitle] = useState("");
  const [size, setSize] = useState("");
  const [color, setColor] = useState("");
  const [priceINR, setPriceINR] = useState(0);
  const [onHand, setOnHand] = useState(0);

  return (
    <form
      className="grid gap-3 sm:grid-cols-2"
      onSubmit={(e) => {
        e.preventDefault();
        startTransition(async () => {
          try {
            setError(null);
            await adminCreateVariant(token, productId, {
              sku,
              title,
              size,
              color,
              price_inr: priceINR,
              on_hand: onHand
            });
            setSku("");
            setTitle("");
            setSize("");
            setColor("");
            setPriceINR(0);
            setOnHand(0);
            await onCreated();
          } catch (e) {
            setError(e instanceof Error ? e.message : "Create variant failed");
          }
        });
      }}
    >
      <label className="space-y-1">
        <div className="text-sm font-medium">SKU</div>
        <input
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          value={sku}
          onChange={(e) => setSku(e.target.value)}
          required
        />
      </label>
      <label className="space-y-1">
        <div className="text-sm font-medium">Title</div>
        <input
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />
      </label>
      <label className="space-y-1">
        <div className="text-sm font-medium">Size</div>
        <input
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          value={size}
          onChange={(e) => setSize(e.target.value)}
        />
      </label>
      <label className="space-y-1">
        <div className="text-sm font-medium">Color</div>
        <input
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          value={color}
          onChange={(e) => setColor(e.target.value)}
        />
      </label>
      <label className="space-y-1">
        <div className="text-sm font-medium">Price (paise)</div>
        <input
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          type="number"
          min={0}
          value={priceINR}
          onChange={(e) => setPriceINR(Number(e.target.value))}
          required
        />
      </label>
      <label className="space-y-1">
        <div className="text-sm font-medium">On hand</div>
        <input
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          type="number"
          min={0}
          value={onHand}
          onChange={(e) => setOnHand(Number(e.target.value))}
          required
        />
      </label>
      <div className="sm:col-span-2">
        <button
          className="rounded-lg bg-slate-900 px-3 py-2 text-sm text-white disabled:opacity-50"
          disabled={isPending}
          type="submit"
        >
          {isPending ? "Creating…" : "Add variant"}
        </button>
      </div>
    </form>
  );
}

