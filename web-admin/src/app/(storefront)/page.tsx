import Link from "next/link";

export default function HomePage() {
  return (
    <div className="animate-fade-in">
      {/* Hero */}
      <section className="relative overflow-hidden bg-vexo-black px-4 py-24 sm:px-6 lg:px-8">
        <div className="absolute inset-0 bg-[linear-gradient(135deg,var(--vexo-teal)_0%,transparent_50%)] opacity-20" />
        <div className="relative mx-auto max-w-7xl text-center">
          <p className="mb-4 text-sm font-medium uppercase tracking-[0.2em] text-vexo-teal">
            Streetwear & Activewear
          </p>
          <h1 className="text-4xl font-bold tracking-tight text-white sm:text-5xl lg:text-6xl">
            Style that moves
            <br />
            <span className="text-vexo-teal">with you</span>
          </h1>
          <p className="mx-auto mt-6 max-w-xl text-lg text-vexo-gray">
            Futuristic eCommerce experience. Discover curated outfits designed for
            the modern lifestyle.
          </p>
          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              href="/catalog"
              className="w-full rounded-lg bg-vexo-teal px-8 py-3.5 text-center font-medium text-white transition-all hover:bg-vexo-teal/90 hover:shadow-lg sm:w-auto"
            >
              Shop Now
            </Link>
            <Link
              href="/admin/login"
              className="w-full rounded-lg border border-vexo-gray/50 px-8 py-3.5 text-center font-medium text-white transition-all hover:bg-white/10 sm:w-auto"
            >
              Admin
            </Link>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <div className="grid gap-8 sm:grid-cols-2 lg:grid-cols-3">
          <Link
            href="/catalog"
            className="group animate-slide-up rounded-2xl border border-vexo-gray/50 bg-white p-6 transition-all hover:border-vexo-teal/50 hover:shadow-xl"
            style={{ animationDelay: "0.1s" }}
          >
            <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-vexo-teal/20 text-vexo-teal transition-colors group-hover:bg-vexo-teal/30">
              <span className="text-xl">✦</span>
            </div>
            <h3 className="font-semibold text-vexo-black">Browse Catalog</h3>
            <p className="mt-2 text-sm text-vexo-brown">
              View products and variants. Find your perfect fit.
            </p>
          </Link>
          <div className="rounded-2xl border border-vexo-gray/50 bg-white p-6">
            <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-vexo-taupe/20 text-vexo-taupe">
              <span className="text-xl">✦</span>
            </div>
            <h3 className="font-semibold text-vexo-black">Curated Collection</h3>
            <p className="mt-2 text-sm text-vexo-brown">
              Handpicked streetwear and activewear for every occasion.
            </p>
          </div>
          <div className="rounded-2xl border border-vexo-gray/50 bg-white p-6">
            <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-vexo-brown/20 text-vexo-brown">
              <span className="text-xl">✦</span>
            </div>
            <h3 className="font-semibold text-vexo-black">Secure Checkout</h3>
            <p className="mt-2 text-sm text-vexo-brown">
              Safe payment with Razorpay. Fast, reliable delivery.
            </p>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="mx-auto max-w-7xl px-4 pb-24 sm:px-6 lg:px-8">
        <div className="rounded-2xl bg-vexo-brown/10 px-6 py-12 text-center sm:px-12">
          <h2 className="text-2xl font-semibold text-vexo-black">
            Ready to explore?
          </h2>
          <p className="mt-2 text-vexo-brown">
            Start shopping our latest collection.
          </p>
          <Link
            href="/catalog"
            className="mt-6 inline-block rounded-lg bg-vexo-black px-6 py-2.5 font-medium text-white transition-colors hover:bg-vexo-brown"
          >
            View Catalog
          </Link>
        </div>
      </section>
    </div>
  );
}
