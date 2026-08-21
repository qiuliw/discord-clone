"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { useAuth } from "@/components/providers/auth-provider";
import { Button } from "@/components/ui/button";
import { login } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const { setUser } = useAuth();
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function onSubmit(formData: FormData) {
    setError("");
    setPending(true);
    try {
      const data = await login({
        email: String(formData.get("email") ?? ""),
        password: String(formData.get("password") ?? ""),
      });
      setUser(data.user);
      router.replace("/");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "sign in failed");
    } finally {
      setPending(false);
    }
  }

  return (
    <form
      className="w-full max-w-sm space-y-4 rounded-xl border border-border bg-card p-6 shadow-sm"
      action={(formData) => void onSubmit(formData)}
    >
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Sign in</h1>
        <p className="text-sm text-muted-foreground">
          Use the account stored on this machine.
        </p>
      </div>
      <label className="block space-y-1.5 text-sm">
        <span>Email</span>
        <input
          name="email"
          type="email"
          required
          autoComplete="email"
          className="h-9 w-full rounded-lg border border-input bg-background px-3 outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
        />
      </label>
      <label className="block space-y-1.5 text-sm">
        <span>Password</span>
        <input
          name="password"
          type="password"
          required
          autoComplete="current-password"
          className="h-9 w-full rounded-lg border border-input bg-background px-3 outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
        />
      </label>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <Button type="submit" className="w-full" disabled={pending}>
        {pending ? "Signing in..." : "Sign in"}
      </Button>
      <p className="text-center text-sm text-muted-foreground">
        No account?{" "}
        <Link href="/register" className="text-foreground underline-offset-4 hover:underline">
          Sign up
        </Link>
      </p>
    </form>
  );
}
