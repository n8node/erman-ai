import Link from "next/link";

export default function PublicPagesLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-bg3">
      <header className="border-b border-border bg-bg">
        <div className="mx-auto flex max-w-3xl items-center justify-between px-6 py-4">
          <Link href="/" className="flex items-center gap-2.5">
            <div className="flex h-7 w-7 items-center justify-center rounded-md bg-text text-sm font-semibold text-white">
              E
            </div>
            <span className="text-sm font-medium">Erman AI</span>
          </Link>
          <Link href="/" className="text-sm text-text2 hover:text-text">
            ← На главную
          </Link>
        </div>
      </header>
      <main className="mx-auto max-w-3xl px-6 py-10">{children}</main>
    </div>
  );
}
