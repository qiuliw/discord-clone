export type User = {
  id: number;
  email: string;
  name: string;
};

type ErrorBody = {
  error?: string;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    credentials: "include",
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  });

  const data = (await res.json().catch(() => ({}))) as T & ErrorBody;
  if (!res.ok) {
    throw new Error(data.error || "request failed");
  }
  return data;
}

export function getMe() {
  return request<{ user: User }>("/api/auth/me");
}

export function register(input: {
  email: string;
  password: string;
  name: string;
}) {
  return request<{ user: User }>("/api/auth/register", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function login(input: { email: string; password: string }) {
  return request<{ user: User }>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function logout() {
  return request<{ ok: boolean }>("/api/auth/logout", { method: "POST" });
}
