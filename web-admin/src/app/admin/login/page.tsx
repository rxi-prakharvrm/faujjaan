import AdminLoginClient from "./login-client";

export default function AdminLoginPage() {
  return (
    <main className="space-y-6">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold">Admin Login</h1>
        <p className="text-slate-600">
          Use the seeded dev credentials: <span className="font-mono">admin@example.com</span>{" "}
          / <span className="font-mono">admin12345</span>
        </p>
      </header>
      <AdminLoginClient />
    </main>
  );
}

