const DEFAULT_API_BASE_URL = "http://localhost:8081";

export type Variant = {
  id: string;
  product_id: string;
  sku: string;
  title: string;
  size: string;
  color: string;
  price_inr: number;
  compare_at_price_inr?: number | null;
  on_hand: number;
  reserved: number;
};

export type Product = {
  id: string;
  slug: string;
  name: string;
  description: string;
  status: string;
  variants: Variant[];
};

export type CartItem = {
  variant_id: string;
  sku: string;
  product_name: string;
  variant_title: string;
  unit_price_inr: number;
  quantity: number;
  line_total_inr: number;
};

export type Cart = {
  id: string;
  status: string;
  items: CartItem[];
  subtotal_inr: number;
};

export type CheckoutResult = {
  order_id: string;
  payment_id: string;
  amount_inr: number;
  currency: string;
  provider: string;
  razorpay?: {
    key_id: string;
    order_id: string;
    amount_inr: number;
    currency: string;
  } | null;
};

function apiBaseUrl(): string {
  if (typeof window === "undefined") {
    return (
      process.env.API_SERVER_BASE_URL ??
      process.env.NEXT_PUBLIC_API_BASE_URL ??
      DEFAULT_API_BASE_URL
    );
  }
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? DEFAULT_API_BASE_URL;
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const url = `${apiBaseUrl()}${path}`;
  const res = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    },
    cache: "no-store"
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`API ${res.status}: ${text || res.statusText}`);
  }
  return (await res.json()) as T;
}

export async function fetchProducts(): Promise<Product[]> {
  const res = await apiFetch<{ products: Product[] }>("/v1/products");
  return res.products;
}

export async function fetchProduct(slug: string): Promise<Product> {
  return await apiFetch<Product>(`/v1/products/${encodeURIComponent(slug)}`);
}

export async function createCart(): Promise<string> {
  const res = await apiFetch<{ cart_id: string }>("/v1/cart", {
    method: "POST"
  });
  return res.cart_id;
}

export async function fetchCart(cartId: string): Promise<Cart> {
  return await apiFetch<Cart>(`/v1/cart/${cartId}`);
}

export async function upsertCartItem(
  cartId: string,
  variantId: string,
  quantity: number
): Promise<void> {
  await apiFetch(`/v1/cart/${cartId}/items`, {
    method: "POST",
    body: JSON.stringify({ variant_id: variantId, quantity })
  });
}

export async function deleteCartItem(
  cartId: string,
  variantId: string
): Promise<void> {
  await apiFetch(`/v1/cart/${cartId}/items/${variantId}`, {
    method: "DELETE"
  });
}

export async function checkoutFromCart(input: {
  cart_id: string;
  customer_name: string;
  customer_phone: string;
  customer_email?: string;
  shipping_address: Record<string, unknown>;
}): Promise<CheckoutResult> {
  return await apiFetch<CheckoutResult>("/v1/checkout", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function verifyRazorpayPayment(input: {
  razorpay_order_id: string;
  razorpay_payment_id: string;
  razorpay_signature: string;
}): Promise<void> {
  await apiFetch("/v1/payments/razorpay/verify", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

