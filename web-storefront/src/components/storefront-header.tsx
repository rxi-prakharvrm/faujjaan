"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";

export default function StorefrontHeader() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const pathname = usePathname();

  useEffect(() => {
    const token = localStorage.getItem("customer_token");
    setIsLoggedIn(!!token);
  }, [pathname]);

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
          {isLoggedIn ? (
            <Link
              href="/account"
              className="text-sm font-medium text-vexo-teal transition-colors hover:text-vexo-brown"
            >
              Account
            </Link>
          ) : (
            <Link
              href="/login"
              className="text-sm font-medium text-vexo-teal transition-colors hover:text-vexo-brown"
            >
              Login
            </Link>
          )}
        </nav>
      </div>
    </header>
  );
}
