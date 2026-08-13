"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth-provider";
import { Button, buttonVariants } from "@/components/ui/button";
import { logout } from "@/lib/api";
import { cn } from "@/lib/utils";

export function SiteHeader() {
  const { user, loading, setUser } = useAuth();
  const router = useRouter();

  async function handleLogout() {
    await logout();
    setUser(null);
    router.push("/");
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
            <Button variant="outline" size="sm" onClick={() => void handleLogout()}>
              Log out
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
