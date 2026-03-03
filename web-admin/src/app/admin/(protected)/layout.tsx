import type { ReactNode } from "react";
import AdminShell from "./shell";

export default function AdminProtectedLayout({ children }: { children: ReactNode }) {
  return <AdminShell>{children}</AdminShell>;
}

