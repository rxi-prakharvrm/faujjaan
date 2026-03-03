import Link from "next/link";
import { fetchProduct } from "@/lib/api";
import ProductClient from "./product-client";

export default async function ProductPage({
  params
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const product = await fetchProduct(slug);

  return (
    <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <nav className="mb-8 flex items-center gap-2 text-sm">
        <Link
          href="/catalog"
          className="text-vexo-brown transition-colors hover:text-vexo-black"
        >
          ← Back to catalog
        </Link>
      </nav>

      <div className="grid gap-12 lg:grid-cols-2">
        <div className="animate-fade-in">
          <div className="aspect-square rounded-2xl bg-vexo-gray/30 flex items-center justify-center">
            <span className="text-8xl font-light text-vexo-teal/50">✦</span>
          </div>
        </div>

        <div className="animate-slide-up space-y-6">
          <div>
            <p className="text-sm font-medium uppercase tracking-[0.2em] text-vexo-teal">
              Product
            </p>
            <h1 className="mt-2 text-3xl font-bold tracking-tight text-vexo-black">
              {product.name}
            </h1>
            <p className="mt-3 text-vexo-brown">{product.description}</p>
          </div>

          <ProductClient product={product} />
        </div>
      </div>
    </div>
  );
}
