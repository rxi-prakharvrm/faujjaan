"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { login } from "@/lib/api";

export default function LoginPage() {
    const router = useRouter();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [isPending, setIsPending] = useState(false);

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        setError(null);
        setIsPending(true);

        try {
            const res = await login({ email, password });
            localStorage.setItem("customer_token", res.token);
            router.push("/");
        } catch (err) {
            setError(err instanceof Error ? err.message : "Login failed");
        } finally {
            setIsPending(false);
        }
    }

    return (
        <main className="mx-auto max-w-md px-4 py-24 sm:px-6 lg:px-8">
            <div className="rounded-2xl border border-vexo-gray/50 bg-white p-8 shadow-sm">
                <h1 className="text-2xl font-bold tracking-tight text-vexo-black">
                    Welcome back
                </h1>
                <p className="mt-2 text-sm text-vexo-brown">
                    Enter your credentials to access your account.
                </p>

                <form onSubmit={handleSubmit} className="mt-8 space-y-6">
                    <div className="space-y-4">
                        <div>
                            <label
                                htmlFor="email"
                                className="block text-sm font-medium text-vexo-black"
                            >
                                Email address
                            </label>
                            <input
                                id="email"
                                type="email"
                                required
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                className="mt-1 block w-full rounded-lg border border-vexo-gray/50 px-3 py-2 text-vexo-black focus:border-vexo-teal focus:ring-vexo-teal sm:text-sm"
                            />
                        </div>

                        <div>
                            <label
                                htmlFor="password"
                                className="block text-sm font-medium text-vexo-black"
                            >
                                Password
                            </label>
                            <input
                                id="password"
                                type="password"
                                required
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="mt-1 block w-full rounded-lg border border-vexo-gray/50 px-3 py-2 text-vexo-black focus:border-vexo-teal focus:ring-vexo-teal sm:text-sm"
                            />
                        </div>
                    </div>

                    {error && (
                        <div className="rounded-lg bg-rose-50 p-3 text-sm text-rose-800">
                            {error}
                        </div>
                    )}

                    <button
                        type="submit"
                        disabled={isPending}
                        className="w-full rounded-lg bg-vexo-black py-3 font-medium text-white transition-all hover:bg-vexo-brown disabled:opacity-50"
                    >
                        {isPending ? "Signing in..." : "Sign in"}
                    </button>
                </form>

                <p className="mt-6 text-center text-sm text-vexo-brown">
                    Don&apos;t have an account?{" "}
                    <Link href="/register" className="font-medium text-vexo-teal hover:underline">
                        Create one
                    </Link>
                </p>
            </div>
        </main>
    );
}
