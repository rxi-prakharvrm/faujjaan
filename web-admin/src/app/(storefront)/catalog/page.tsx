import Link from "next/link";
import { fetchProducts } from "@/lib/api";
import { formatINR } from "@/lib/money";

export default async function CatalogPage() {
  const products = await fetchProducts();

  return (
    <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div className="mb-12 animate-fade-in">
        <p className="text-sm font-medium uppercase tracking-[0.2em] text-vexo-teal">
          Collection
        </p>
        <h1 className="mt-2 text-3xl font-bold tracking-tight text-vexo-black sm:text-4xl">
          Shop All
        </h1>
        <p className="mt-2 text-vexo-brown">
          {products.length} product{products.length === 1 ? "" : "s"}
        </p>
      </div>

      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {products.map((p, i) => {
          const minPrice = Math.min(...p.variants.map((v) => v.price_inr));
          const inStock = p.variants.some((v) => v.on_hand - v.reserved > 0);
          return (
            <Link
              key={p.id}
              href={`/product/${p.slug}`}
              className="group animate-slide-up rounded-2xl border border-vexo-gray/50 bg-white p-6 transition-all hover:border-vexo-teal/50 hover:shadow-xl"
              style={{ animationDelay: `${i * 0.05}s` }}
            >
              <div className="mb-4 flex h-48 items-center justify-center rounded-xl bg-vexo-gray/30 text-vexo-teal/50 transition-colors group-hover:bg-vexo-teal/20">
                <span className="text-4xl font-light">✦</span>
              </div>
              <div className="space-y-2">
                <h3 className="font-semibold text-vexo-black group-hover:text-vexo-teal">
                  {p.name}
                </h3>
                <p className="line-clamp-2 text-sm text-vexo-brown">
                  {p.description}
                </p>
                <div className="flex items-center justify-between pt-2">
                  <span className="font-medium text-vexo-black">
                    {formatINR(minPrice)}
                  </span>
                  <span
                    className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      inStock
                        ? "bg-emerald-100 text-emerald-800"
                        : "bg-rose-100 text-rose-800"
                    }`}
                  >
                    {inStock ? "In stock" : "Out of stock"}
                  </span>
                </div>
              </div>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
