const CART_ID_KEY = "clothes_shop_cart_id";

export function getStoredCartId(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(CART_ID_KEY);
}

export function setStoredCartId(cartId: string) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(CART_ID_KEY, cartId);
}

export function clearStoredCartId() {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(CART_ID_KEY);
}

