import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Vexo | Streetwear & Activewear",
  description: "Stylish outfits for streetwear and activewear. Futuristic eCommerce experience."
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-vexo-gray/30 font-sans text-vexo-black antialiased">
        {children}
      </body>
    </html>
  );
}

