"use client";

import Link from "next/link";
import { useEffect, useMemo, useState, useTransition } from "react";
import type { Cart, CheckoutResult } from "@/lib/api";
import {
  checkoutFromCart,
  deleteCartItem,
  fetchCart,
  upsertCartItem,
  verifyRazorpayPayment
} from "@/lib/api";
import {
  clearStoredCartId,
  getStoredCartId
} from "@/lib/cart";
import { formatINR } from "@/lib/money";

type Status = "idle" | "loading" | "error" | "ready";

export default function CartClient() {
  const [status, setStatus] = useState<Status>("idle");
  const [cartId, setCartId] = useState<string | null>(null);
  const [cart, setCart] = useState<Cart | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();
  const [checkoutRes, setCheckoutRes] = useState<CheckoutResult | null>(null);

  const total = useMemo(() => cart?.subtotal_inr ?? 0, [cart]);

  async function refresh(cid: string) {
    const c = await fetchCart(cid);
    setCart(c);
  }

  useEffect(() => {
    const cid = getStoredCartId();
    setCartId(cid);
    if (!cid) {
      setStatus("ready");
      return;
    }
    setStatus("loading");
    refresh(cid)
      .then(() => setStatus("ready"))
      .catch((e) => {
        setError(e instanceof Error ? e.message : "Failed to load cart");
        setStatus("error");
      });
  }, []);

  async function updateQty(variantId: string, quantity: number) {
    if (!cartId) return;
    startTransition(async () => {
      try {
        await upsertCartItem(cartId, variantId, quantity);
        await refresh(cartId);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to update item");
      }
    });
  }

  async function removeItem(variantId: string) {
    if (!cartId) return;
    startTransition(async () => {
      try {
        await deleteCartItem(cartId, variantId);
        await refresh(cartId);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to remove item");
      }
    });
  }

  async function submitCheckout(form: HTMLFormElement) {
    if (!cartId) return;
    const fd = new FormData(form);
    const customer_name = String(fd.get("name") ?? "").trim();
    const customer_phone = String(fd.get("phone") ?? "").trim();
    const customer_email = String(fd.get("email") ?? "").trim();
    const address1 = String(fd.get("address1") ?? "").trim();
    const city = String(fd.get("city") ?? "").trim();
    const state = String(fd.get("state") ?? "").trim();
    const pincode = String(fd.get("pincode") ?? "").trim();

    if (!customer_name || !customer_phone || !address1 || !city || !state || !pincode) {
      setError("Please fill name/phone and full address.");
      return;
    }

    startTransition(async () => {
      try {
        const res = await checkoutFromCart({
          cart_id: cartId,
          customer_name,
          customer_phone,
          customer_email: customer_email || undefined,
          shipping_address: { address1, city, state, pincode }
        });
        if (res.razorpay?.order_id && res.razorpay?.key_id) {
          await openRazorpayCheckout({
            keyId: res.razorpay.key_id,
            amount: res.razorpay.amount_inr,
            currency: res.razorpay.currency,
            razorpayOrderId: res.razorpay.order_id,
            customer: { name: customer_name, phone: customer_phone, email: customer_email }
          });
        }
        setCheckoutRes(res);
        clearStoredCartId();
        setCartId(null);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Checkout failed");
      }
    });
  }

  if (checkoutRes) {
    return (
      <div className="animate-fade-in rounded-2xl border border-emerald-200 bg-emerald-50 p-6 space-y-4">
        <div className="font-semibold text-emerald-900">Order created</div>
        <div className="text-sm text-emerald-900">
          Order: <span className="font-mono">{checkoutRes.order_id}</span>
        </div>
        <div className="text-sm text-emerald-900">
          Amount: <span className="font-medium">{formatINR(checkoutRes.amount_inr)}</span>
        </div>
        <div className="text-sm text-emerald-900">
          Payment provider: {checkoutRes.provider}
          {checkoutRes.razorpay?.order_id ? (
            <>
              {" "}
              — Razorpay order{" "}
              <span className="font-mono text-xs">{checkoutRes.razorpay.order_id}</span>
            </>
          ) : null}
        </div>
        <Link
          className="inline-block rounded-lg bg-vexo-black px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-vexo-brown"
          href="/catalog"
        >
          Continue shopping
        </Link>
      </div>
    );
  }

  if (status === "loading") {
    return (
      <div className="rounded-2xl border border-vexo-gray/50 bg-white p-12 text-center text-vexo-brown">
        Loading cart…
      </div>
    );
  }

  if (!cartId) {
    return (
      <div className="animate-fade-in rounded-2xl border border-vexo-gray/50 bg-white p-12 text-center">
        <p className="font-medium text-vexo-black">Your cart is empty.</p>
        <Link
          href="/catalog"
          className="mt-4 inline-block rounded-lg bg-vexo-teal px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-vexo-teal/90"
        >
          Browse products
        </Link>
      </div>
    );
  }

  if (status === "error") {
    return (
      <div className="rounded-2xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-800">
        {error ?? "Failed to load cart"}
      </div>
    );
  }

  if (!cart) return null;

  return (
    <div className="grid gap-8 lg:grid-cols-3">
      <section className="lg:col-span-2 space-y-4">
        {cart.items.length === 0 ? (
          <div className="rounded-2xl border border-vexo-gray/50 bg-white p-12 text-center">
            Your cart is empty.
          </div>
        ) : (
          cart.items.map((it) => (
            <div
              key={it.variant_id}
              className="rounded-2xl border border-vexo-gray/50 bg-white p-6 space-y-4"
            >
              <div className="flex items-start justify-between gap-4">
                <div>
                  <div className="font-semibold text-vexo-black">{it.product_name}</div>
                  <div className="text-sm text-vexo-brown">{it.variant_title}</div>
                  <div className="mt-1 text-sm font-medium text-vexo-teal">
                    {formatINR(it.unit_price_inr)}
                  </div>
                </div>
                <button
                  className="text-sm font-medium text-vexo-brown transition-colors hover:text-rose-600 disabled:opacity-50"
                  disabled={isPending}
                  onClick={() => removeItem(it.variant_id)}
                >
                  Remove
                </button>
              </div>

              <div className="flex items-center justify-between gap-4 border-t border-vexo-gray/30 pt-4">
                <div className="flex items-center gap-3">
                  <label className="text-sm text-vexo-brown" htmlFor={`qty-${it.variant_id}`}>
                    Qty
                  </label>
                  <input
                    id={`qty-${it.variant_id}`}
                    className="w-20 rounded-lg border border-vexo-gray/80 px-3 py-2 text-vexo-black focus:border-vexo-teal focus:outline-none focus:ring-2 focus:ring-vexo-teal/20"
                    type="number"
                    min={1}
                    max={20}
                    value={it.quantity}
                    disabled={isPending}
                    onChange={(e) =>
                      updateQty(it.variant_id, Number(e.target.value))
                    }
                  />
                </div>
                <div className="font-semibold text-vexo-black">
                  {formatINR(it.line_total_inr)}
                </div>
              </div>
            </div>
          ))
        )}
      </section>

      <aside className="space-y-6">
        <div className="rounded-2xl border border-vexo-gray/50 bg-white p-6 space-y-2">
          <div className="flex items-center justify-between">
            <div className="text-sm text-vexo-brown">Subtotal</div>
            <div className="font-semibold text-vexo-black">{formatINR(total)}</div>
          </div>
          <p className="text-xs text-vexo-brown/80">
            Shipping/tax will be applied by the API (MVP defaults to 0).
          </p>
        </div>

        <form
          className="rounded-2xl border border-vexo-gray/50 bg-white p-6 space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            void submitCheckout(e.currentTarget);
          }}
        >
          <div className="font-semibold text-vexo-black">Shipping & Checkout</div>

          <div className="grid gap-3">
            <input
              className="rounded-lg border border-vexo-gray/80 px-4 py-2.5 text-vexo-black placeholder:text-vexo-brown/60 focus:border-vexo-teal focus:outline-none focus:ring-2 focus:ring-vexo-teal/20"
              name="name"
              placeholder="Name"
            />
            <input
              className="rounded-lg border border-vexo-gray/80 px-4 py-2.5 text-vexo-black placeholder:text-vexo-brown/60 focus:border-vexo-teal focus:outline-none focus:ring-2 focus:ring-vexo-teal/20"
              name="phone"
              placeholder="Phone"
            />
            <input
              className="rounded-lg border border-vexo-gray/80 px-4 py-2.5 text-vexo-black placeholder:text-vexo-brown/60 focus:border-vexo-teal focus:outline-none focus:ring-2 focus:ring-vexo-teal/20"
              name="email"
              placeholder="Email (optional)"
            />
            <input
              className="rounded-lg border border-vexo-gray/80 px-4 py-2.5 text-vexo-black placeholder:text-vexo-brown/60 focus:border-vexo-teal focus:outline-none focus:ring-2 focus:ring-vexo-teal/20"
              name="address1"
              placeholder="Address line"
            />
            <div className="grid gap-3 sm:grid-cols-3">
              <input
                className="rounded-lg border border-vexo-gray/80 px-4 py-2.5 text-vexo-black placeholder:text-vexo-brown/60 focus:border-vexo-teal focus:outline-none focus:ring-2 focus:ring-vexo-teal/20"
                name="city"
                placeholder="City"
              />
              <input
                className="rounded-lg border border-vexo-gray/80 px-4 py-2.5 text-vexo-black placeholder:text-vexo-brown/60 focus:border-vexo-teal focus:outline-none focus:ring-2 focus:ring-vexo-teal/20"
                name="state"
                placeholder="State"
              />
              <input
                className="rounded-lg border border-vexo-gray/80 px-4 py-2.5 text-vexo-black placeholder:text-vexo-brown/60 focus:border-vexo-teal focus:outline-none focus:ring-2 focus:ring-vexo-teal/20"
                name="pincode"
                placeholder="Pincode"
              />
            </div>
          </div>

          <button
            className="w-full rounded-lg bg-vexo-black px-4 py-3.5 font-medium text-white transition-all hover:bg-vexo-brown disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={isPending || cart.items.length === 0}
            type="submit"
          >
            {isPending ? "Creating order…" : "Proceed to payment"}
          </button>

          {error ? (
            <div className="rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
              {error}
            </div>
          ) : null}
        </form>
      </aside>
    </div>
  );
}

async function openRazorpayCheckout(input: {
  keyId: string;
  amount: number;
  currency: string;
  razorpayOrderId: string;
  customer: { name: string; phone: string; email: string };
}) {
  await loadRazorpayScript();
  type RazorpayPaymentResponse = {
    razorpay_order_id: string;
    razorpay_payment_id: string;
    razorpay_signature: string;
  };
  type RazorpayOptions = {
    key: string;
    amount: number;
    currency: string;
    name: string;
    order_id: string;
    prefill: { name: string; email: string; contact: string };
    handler: (resp: RazorpayPaymentResponse) => void | Promise<void>;
    modal: { ondismiss: () => void };
  };
  type RazorpayInstance = { open: () => void };
  type RazorpayConstructor = new (opts: RazorpayOptions) => RazorpayInstance;

  const RazorpayCtor = (window as unknown as { Razorpay?: RazorpayConstructor })
    .Razorpay;
  if (!RazorpayCtor) throw new Error("Razorpay SDK not available");

  await new Promise<void>((resolve, reject) => {
    const rzp = new RazorpayCtor({
      key: input.keyId,
      amount: input.amount,
      currency: input.currency,
      name: "Vexo",
      order_id: input.razorpayOrderId,
      prefill: {
        name: input.customer.name,
        email: input.customer.email,
        contact: input.customer.phone
      },
      handler: async (resp) => {
        try {
          await verifyRazorpayPayment({
            razorpay_order_id: resp.razorpay_order_id,
            razorpay_payment_id: resp.razorpay_payment_id,
            razorpay_signature: resp.razorpay_signature
          });
          resolve();
        } catch (e) {
          reject(e);
        }
      },
      modal: {
        ondismiss: () => reject(new Error("Payment cancelled"))
      }
    });
    rzp.open();
  });
}

function loadRazorpayScript(): Promise<void> {
  if (typeof window === "undefined") return Promise.resolve();
  const existing = document.querySelector<HTMLScriptElement>(
    'script[src="https://checkout.razorpay.com/v1/checkout.js"]'
  );
  if (existing) return Promise.resolve();
  return new Promise((resolve, reject) => {
    const s = document.createElement("script");
    s.src = "https://checkout.razorpay.com/v1/checkout.js";
    s.async = true;
    s.onload = () => resolve();
    s.onerror = () => reject(new Error("Failed to load Razorpay SDK"));
    document.body.appendChild(s);
  });
}
