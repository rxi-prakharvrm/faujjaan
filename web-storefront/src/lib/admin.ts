const TOKEN_KEY = "clothes_shop_admin_token";

export function getAdminToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setAdminToken(token: string) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(TOKEN_KEY, token);
}

export function clearAdminToken() {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(TOKEN_KEY);
}

function apiBaseUrl(): string {
  if (typeof window === "undefined") {
    return (
      process.env.API_SERVER_BASE_URL ??
      process.env.NEXT_PUBLIC_API_BASE_URL ??
      "http://localhost:8081"
    );
  }
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8081";
}

async function apiFetch<T>(
  path: string,
  init?: RequestInit & { token?: string }
): Promise<T> {
  const url = `${apiBaseUrl()}${path}`;
  const headers: Record<string, string> = {
    "Content-Type": "application/json"
  };
  if (init?.token) headers["Authorization"] = `Bearer ${init.token}`;

  const res = await fetch(url, {
    ...init,
    headers: { ...headers, ...(init?.headers ?? {}) },
    cache: "no-store"
  });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`API ${res.status}: ${text || res.statusText}`);
  }
  return (await res.json()) as T;
}

export async function adminLogin(email: string, password: string) {
  const res = await apiFetch<{ token: string; role: string }>(
    "/v1/admin/login",
    {
      method: "POST",
      body: JSON.stringify({ email, password })
    }
  );
  return res.token;
}

export type AdminProduct = {
  id: string;
  slug: string;
  name: string;
  description: string;
  status: string;
  variants: Array<{
    id: string;
    sku: string;
    title: string;
    size: string;
    color: string;
    price_inr: number;
    on_hand: number;
    reserved: number;
  }>;
};

export async function adminListProducts(token: string) {
  const res = await apiFetch<{ products: AdminProduct[] }>("/v1/admin/products", {
    token
  });
  return res.products;
}

export async function adminCreateProduct(
  token: string,
  input: {
    slug: string;
    name: string;
    description: string;
    status: string;
    variants: Array<{
      sku: string;
      title: string;
      size: string;
      color: string;
      price_inr: number;
      on_hand: number;
    }>;
  }
) {
  return await apiFetch<AdminProduct>("/v1/admin/products", {
    token,
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function adminUpdateProduct(
  token: string,
  productId: string,
  input: { slug: string; name: string; description: string; status: string }
) {
  await apiFetch(`/v1/admin/products/${productId}`, {
    token,
    method: "PUT",
    body: JSON.stringify(input)
  });
}

export async function adminCreateVariant(
  token: string,
  productId: string,
  input: {
    sku: string;
    title: string;
    size: string;
    color: string;
    price_inr: number;
    on_hand: number;
  }
) {
  return await apiFetch(`/v1/admin/products/${productId}/variants`, {
    token,
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function adminUpdateVariant(
  token: string,
  variantId: string,
  input: { title: string; size: string; color: string; price_inr: number }
) {
  await apiFetch(`/v1/admin/variants/${variantId}`, {
    token,
    method: "PUT",
    body: JSON.stringify(input)
  });
}

export async function adminAdjustInventory(
  token: string,
  input: { variant_id: string; delta: number }
) {
  return await apiFetch<{ on_hand: number; reserved: number }>(
    "/v1/admin/inventory/adjust",
    {
      token,
      method: "POST",
      body: JSON.stringify(input)
    }
  );
}

export type AdminOrderSummary = {
  id: string;
  status: string;
  total_inr: number;
  created_at: string;
};

export async function adminListOrders(token: string) {
  const res = await apiFetch<{ orders: AdminOrderSummary[] }>("/v1/admin/orders", {
    token
  });
  return res.orders;
}

export async function adminGetOrder(token: string, orderId: string) {
  return await apiFetch(`/v1/admin/orders/${orderId}`, { token });
}

