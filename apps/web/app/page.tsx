"use client";

import Link from "next/link";

import { useAuth } from "@/components/auth-provider";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export default function Home() {
  const { user, loading } = useAuth();

  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-3 px-6 py-16 text-center">
      <h1 className="text-2xl font-semibold tracking-tight">Welcome to Team Chat</h1>
      {loading ? null : user ? (
        <p className="max-w-md text-sm text-muted-foreground">
          Signed in as {user.name} ({user.email}). Your account is stored in the
          local Go API.
        </p>
      ) : (
        <>
          <p className="max-w-md text-sm text-muted-foreground">
            Create an account to get started. Auth runs on the local Go backend,
            not a third-party service.
          </p>
          <div className="flex gap-2">
            <Link
              href="/login"
              className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
            >
              Sign in
            </Link>
            <Link href="/register" className={cn(buttonVariants({ size: "sm" }))}>
              Sign up
            </Link>
          </div>
        </>
      )}
    </main>
  );
}
