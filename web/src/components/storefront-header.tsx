import Link from "next/link";

export default function StorefrontHeader() {
  return (
    <header className="sticky top-0 z-50 border-b border-vexo-gray/50 bg-white/90 backdrop-blur-md">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4 sm:px-6 lg:px-8">
        <Link
          href="/"
          className="text-xl font-semibold tracking-tight text-vexo-black transition-colors hover:text-vexo-teal"
        >
          Vexo
        </Link>
        <nav className="flex items-center gap-6">
          <Link
            href="/catalog"
            className="text-sm font-medium text-vexo-brown transition-colors hover:text-vexo-black"
          >
            Shop
          </Link>
          <Link
            href="/cart"
            className="text-sm font-medium text-vexo-brown transition-colors hover:text-vexo-black"
          >
            Cart
          </Link>
          <Link
            href="/admin/login"
            className="text-sm text-vexo-teal transition-colors hover:text-vexo-brown"
          >
            Admin
          </Link>
        </nav>
      </div>
    </header>
  );
}
