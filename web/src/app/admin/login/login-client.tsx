"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { adminLogin, setAdminToken } from "@/lib/admin";

export default function AdminLoginClient() {
  const router = useRouter();
  const [email, setEmail] = useState("admin@example.com");
  const [password, setPassword] = useState("admin12345");
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  return (
    <form
      className="max-w-md rounded-2xl border border-slate-200 p-4 space-y-3"
      onSubmit={(e) => {
        e.preventDefault();
        setError(null);
        startTransition(async () => {
          try {
            const token = await adminLogin(email, password);
            setAdminToken(token);
            router.push("/admin/products");
          } catch (err) {
            setError(err instanceof Error ? err.message : "Login failed");
          }
        });
      }}
    >
      <div className="space-y-1">
        <div className="text-sm font-medium">Email</div>
        <input
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          autoComplete="email"
        />
      </div>

      <div className="space-y-1">
        <div className="text-sm font-medium">Password</div>
        <input
          className="w-full rounded-lg border border-slate-300 px-3 py-2"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          type="password"
          autoComplete="current-password"
        />
      </div>

      <button
        className="w-full rounded-lg bg-slate-900 px-3 py-2 text-white disabled:opacity-50"
        disabled={isPending}
        type="submit"
      >
        {isPending ? "Signing in…" : "Sign in"}
      </button>

      {error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
          {error}
        </div>
      ) : null}
    </form>
  );
}

