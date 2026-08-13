export default function AuthLayout({ children }: { children: React.ReactNode }) {
    return (
        <div className="flex min-h-full flex-col bg-background bg-red-300">
            {children}
        </div>
    );
}
