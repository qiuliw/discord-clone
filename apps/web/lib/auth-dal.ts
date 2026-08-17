import "server-only";

import { cache } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import type { User } from "@/lib/api";

const cookieName = "token";
const apiURL = process.env.API_PROXY_URL ?? "http://127.0.0.1:8080";

export const getCurrentUser = cache(async (): Promise<User | null> => {
  const token = (await cookies()).get(cookieName)?.value;
  if (!token) {
    return null;
  }

  const response = await fetch(`${apiURL}/api/auth/me`, {
    headers: { Cookie: `${cookieName}=${token}` },
    cache: "no-store",
  });

  if (response.status === 401) {
    return null;
  }
  if (!response.ok) {
    throw new Error(`failed to verify session: ${response.status}`);
  }

  const data = (await response.json()) as { user: User };
  return data.user;
});

export async function requireUser(): Promise<User> {
  const user = await getCurrentUser();
  if (!user) {
    redirect("/login");
  }
  return user;
}
