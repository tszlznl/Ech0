import type { Route } from "./+types/home";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Ech0 — Self-hosted microblog" },
    {
      name: "description",
      content:
        "Self-hosted timeline you own. Share posts, optional comments. Lightweight and open source.",
    },
  ];
}

function ArrowIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      width="14"
      height="14"
      viewBox="0 0 256 256"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden
    >
      <path d="M221.66 133.66l-72 72a8 8 0 01-11.32-11.32L196.69 136H40a8 8 0 010-16h156.69l-58.35-58.34a8 8 0 0111.32-11.32l72 72a8 8 0 010 11.32z" />
    </svg>
  );
}

export default function Home() {
  return (
    <div className="min-h-screen bg-app">
      <header className="mx-auto flex w-full max-w-[min(100%,30rem)] items-center justify-between px-5 py-8">
        <a href="/" className="flex items-center gap-2.5 text-sand-12 no-underline">
          <img
            src="/logo.svg"
            alt="Ech0"
            width={28}
            height={28}
            className="size-7 shrink-0 rounded-sm"
          />
          <span className="text-[0.95rem] font-medium">Ech0</span>
        </a>
        <a
          href="https://memo.vaaat.com/"
          className="text-sm font-medium text-sand-11 transition-colors hover:text-sand-12"
        >
          Live demo
        </a>
      </header>

      <main className="mx-auto flex w-full max-w-[min(100%,30rem)] flex-col gap-y-8 px-5 pb-24">
        <section className="flex flex-col items-center gap-8 text-center">
          <div className="w-full overflow-hidden rounded-xl shadow-[0_16px_40px_-12px_rgba(33,32,28,0.12)] ring-1 ring-sand-6/70">
            <img
              src="/screenshot.png"
              alt="Ech0 interface preview"
              className="block w-full align-middle"
              width={1412}
              height={1131}
              loading="eager"
            />
          </div>

          <h1 className="font-serif text-2xl font-normal leading-snug text-sand-12 text-balance">
            <span className="block">Self-hosted microblog.</span>
            <span className="mt-1 block">Own your timeline.</span>
          </h1>

          <div className="font-serif text-[0.9375rem] font-normal leading-[1.28] tracking-[0.01em] text-sand-11 [&>p+p]:mt-0.5">
            <p>Not a team wiki.</p>
            <p>Not a social network.</p>
            <p className="text-sand-12">A public timeline on your server.</p>
          </div>

          <a
            href="https://github.com/lin-snow/Ech0/releases"
            className="inline-flex items-center gap-1.5 rounded-full border border-sand-6 bg-sand-2/70 px-4 py-2 text-[13px] font-medium text-sand-11 shadow-[0_1px_2px_rgba(33,32,28,0.04)] no-underline transition-colors hover:border-sand-11/20 hover:bg-sand-2 hover:text-sand-12"
          >
            <ArrowIcon className="opacity-70" />
            Get started
          </a>

          <p className="text-xs font-medium text-sand-11/90">
            Lightweight · easy to deploy · open source
          </p>
        </section>

        <section className="border-t border-sand-6 pt-8 font-serif text-[1.0625rem] leading-[1.65] text-sand-11">
          <p className="font-normal text-sand-12">
            After capture comes publishing.
          </p>
          <p className="mt-3 text-[0.98em]">
            Ech0 is for what comes after quick notes: one timeline, optional
            comments, data on your box.
          </p>
        </section>

        <section className="font-serif text-[1.0625rem] leading-[1.65] text-sand-11">
          <h2 className="text-[1.08em] font-semibold leading-snug text-sand-12">
            Why Ech0
          </h2>
          <p className="mt-3 text-[0.98em]">
            Skip PKM bloat and team docs—if you want a small, deployable
            microblog, this fits.
          </p>
        </section>

        <section className="font-serif text-[1.0625rem] leading-[1.65] text-sand-11">
          <h2 className="text-[1.08em] font-semibold leading-snug text-sand-12">
            What you get
          </h2>
          <ol className="mt-4 list-decimal space-y-3.5 pl-[1.35rem] text-[0.98em] marker:text-sand-11">
            <li>
              <span className="font-semibold text-sand-12">Post</span>
              {" — "}
              short posts, links, and media from one UI.
            </li>
            <li>
              <span className="font-semibold text-sand-12">Self-host</span>
              {" — "}
              your data stays on your infrastructure.
            </li>
            <li>
              <span className="font-semibold text-sand-12">Connect</span>
              {" — "}
              optional comments and sharing—lightweight, not a full network.
            </li>
          </ol>
        </section>

        <footer className="border-t border-sand-6 pt-10 text-center text-sm text-sand-11">
          <a
            href="https://www.ech0.app/"
            className="font-medium underline-offset-4 transition-colors hover:text-sand-12"
          >
            Docs
          </a>
          <span className="mx-2 text-sand-6">·</span>
          <a
            href="https://github.com/lin-snow/Ech0"
            className="font-medium underline-offset-4 transition-colors hover:text-sand-12"
          >
            GitHub
          </a>
        </footer>
      </main>
    </div>
  );
}
