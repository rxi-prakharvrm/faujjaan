import Link from "next/link";
import CartClient from "./cart-client";

export default function CartPage() {
  return (
    <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <p className="text-sm font-medium uppercase tracking-[0.2em] text-vexo-teal">
            Checkout
          </p>
          <h1 className="mt-2 text-3xl font-bold tracking-tight text-vexo-black">
            Your Cart
          </h1>
        </div>
        <Link
          href="/catalog"
          className="text-sm font-medium text-vexo-brown transition-colors hover:text-vexo-black"
        >
          Continue shopping
        </Link>
      </div>
      <CartClient />
    </div>
  );
}
