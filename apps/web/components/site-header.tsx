"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { Button, buttonVariants } from "@/components/ui/button";
import { logout } from "@/lib/api";
import { cn } from "@/lib/utils";

export function SiteHeader() {
  const { user, loading, setUser } = useAuth();
  const router = useRouter();
  const [loggingOut, setLoggingOut] = useState(false);
  const [error, setError] = useState("");

  async function handleLogout() {
    setError("");
    setLoggingOut(true);
    try {
      await logout();
      setUser(null);
      router.replace("/login");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "log out failed");
    } finally {
      setLoggingOut(false);
    }
  }

  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-border bg-background px-4">
      <Link href="/" className="text-sm font-semibold tracking-tight">
        Team Chat
      </Link>
      <div className="flex items-center gap-2">
        {loading ? null : user ? (
          <>
            <span className="hidden text-sm text-muted-foreground sm:inline">
              {user.email}
            </span>
            {error ? <span className="text-sm text-destructive">{error}</span> : null}
            <Button
              variant="outline"
              size="sm"
              disabled={loggingOut}
              onClick={() => void handleLogout()}
            >
              {loggingOut ? "Logging out..." : "Log out"}
            </Button>
          </>
        ) : (
          <>
            <Link
              href="/login"
              className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
            >
              Sign in
            </Link>
            <Link href="/register" className={cn(buttonVariants({ size: "sm" }))}>
              Sign up
            </Link>
          </>
        )}
      </div>
    </header>
  );
}
