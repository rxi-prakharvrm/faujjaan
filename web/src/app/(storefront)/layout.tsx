import type { ReactNode } from "react";
import StorefrontFooter from "@/components/storefront-footer";
import StorefrontHeader from "@/components/storefront-header";

export default function StorefrontLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <StorefrontHeader />
      <main className="flex-1">{children}</main>
      <StorefrontFooter />
    </div>
  );
}
