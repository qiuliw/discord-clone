import { requireUser } from "@/lib/auth-dal";

export default async function Home() {
  const user = await requireUser();

  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-3 px-6 py-16 text-center">
      <h1 className="text-2xl font-semibold tracking-tight">Welcome to Team Chat</h1>
      <p className="max-w-md text-sm text-muted-foreground">
        Signed in as {user.name} ({user.email}). Your account is stored in the local
        Go API.
      </p>
    </main>
  );
}
