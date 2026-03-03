import Link from "next/link";

export default function StorefrontFooter() {
  return (
    <footer className="mt-auto border-t border-vexo-gray/50 bg-vexo-black text-white">
      <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="flex flex-col items-center justify-between gap-6 sm:flex-row">
          <Link
            href="/"
            className="text-lg font-semibold tracking-tight text-white transition-opacity hover:opacity-80"
          >
            Vexo
          </Link>
          <nav className="flex items-center gap-8">
            <Link
              href="/catalog"
              className="text-sm text-vexo-gray transition-colors hover:text-white"
            >
              Shop
            </Link>
            <Link
              href="/cart"
              className="text-sm text-vexo-gray transition-colors hover:text-white"
            >
              Cart
            </Link>
          </nav>
        </div>
        <p className="mt-8 text-center text-xs text-vexo-teal/80">
          Streetwear & activewear. Futuristic eCommerce experience.
        </p>
      </div>
    </footer>
  );
}
