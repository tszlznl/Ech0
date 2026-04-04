import { Link } from "react-router";
import type { Route } from "./+types/home";
import { absoluteUrl, DEFAULT_DESCRIPTION, SITE_NAME, siteUrl } from "../site";

const PAGE_TITLE = `${SITE_NAME} — Self-hosted microblog & timeline`;
const OG_IMAGE_PATH = "/screenshot.png";
const OG_IMAGE_WIDTH = 1412;
const OG_IMAGE_HEIGHT = 1131;

export function meta(_args: Route.MetaArgs) {
  const canonical = absoluteUrl("/");
  const imageUrl = absoluteUrl(OG_IMAGE_PATH);
  const isHttps = imageUrl.startsWith("https:");
  return [
    { title: PAGE_TITLE },
    { name: "description", content: DEFAULT_DESCRIPTION },
    {
      name: "keywords",
      content:
        "Ech0, microblog, self-hosted, timeline, open source, blog, personal website, RSS alternative, memo",
    },
    { name: "author", content: "Ech0" },
    { name: "application-name", content: SITE_NAME },
    { name: "robots", content: "index, follow" },
    { name: "theme-color", content: "#f6f4f0" },
    { property: "og:type", content: "website" },
    { property: "og:site_name", content: SITE_NAME },
    { property: "og:title", content: PAGE_TITLE },
    { property: "og:description", content: DEFAULT_DESCRIPTION },
    { property: "og:url", content: canonical },
    { property: "og:image", content: imageUrl },
    { property: "og:image:type", content: "image/png" },
    ...(isHttps
      ? [{ property: "og:image:secure_url", content: imageUrl } as const]
      : []),
    { property: "og:image:width", content: String(OG_IMAGE_WIDTH) },
    { property: "og:image:height", content: String(OG_IMAGE_HEIGHT) },
    {
      property: "og:image:alt",
      content: "Ech0 web interface showing a personal timeline feed",
    },
    { property: "og:locale", content: "en_US" },
    { name: "twitter:card", content: "summary_large_image" },
    { name: "twitter:title", content: PAGE_TITLE },
    { name: "twitter:description", content: DEFAULT_DESCRIPTION },
    { name: "twitter:image", content: imageUrl },
    {
      name: "twitter:image:alt",
      content: "Ech0 web interface showing a personal timeline feed",
    },
  ];
}

export const links: Route.LinksFunction = () => [
  { rel: "canonical", href: absoluteUrl("/") },
  { rel: "preload", href: OG_IMAGE_PATH, as: "image" },
];

function LeafIcon({ className }: { className?: string }) {
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
      <path d="M208 40c-32 28-88 24-128 64-40 40-48 96-48 96s56-8 96-48c40-40 36-96 64-128 4-4 12-4 16 0s4 12 0 16zM72 216s24-8 48-32c28-28 32-72 32-72s-44 4-72 32c-24 24-32 48-32 48z" />
    </svg>
  );
}

function buildHomeJsonLd() {
  const base = siteUrl();
  const imageUrl = absoluteUrl(OG_IMAGE_PATH);
  return {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "WebSite",
        "@id": `${base}/#website`,
        url: base,
        name: SITE_NAME,
        description: DEFAULT_DESCRIPTION,
        inLanguage: "en",
        publisher: { "@id": `${base}/#software` },
      },
      {
        "@type": "SoftwareApplication",
        "@id": `${base}/#software`,
        name: SITE_NAME,
        description: DEFAULT_DESCRIPTION,
        url: base,
        image: imageUrl,
        screenshot: imageUrl,
        applicationCategory: "WebApplication",
        operatingSystem: "Linux, Docker, self-hosted",
        license: "https://github.com/lin-snow/Ech0/blob/main/LICENSE",
        codeRepository: "https://github.com/lin-snow/Ech0",
        sameAs: ["https://github.com/lin-snow/Ech0"],
        offers: {
          "@type": "Offer",
          price: "0",
          priceCurrency: "USD",
        },
      },
    ],
  } as const;
}

/** Dashed divider + compact gap to copy (line → text, block → block). */
const dashedSection = "border-t border-dashed border-sand-6 mt-10 pt-5";

export default function Home() {
  return (
    <div className="min-h-screen bg-app">
      <header className="mx-auto flex w-full max-w-[min(100%,30rem)] items-center justify-between px-5 pt-8 pb-4">
        <a
          href="/"
          className="flex items-center gap-2.5 text-sand-12 no-underline"
        >
          <img
            src="/logo.svg"
            alt="Ech0"
            width={28}
            height={28}
            className="size-7 shrink-0 rounded-sm"
          />
          <span className="text-[0.95rem] font-medium tracking-tight">
            Ech0
          </span>
        </a>
        <Link
          to="/docs"
          prefetch="viewport"
          className="text-sm font-medium text-sand-11 transition-colors hover:text-sand-12"
        >
          Docs
        </Link>
      </header>

      <div className="mx-auto w-full max-w-[min(100%,30rem)] px-5">
        <div className="overflow-hidden rounded-xl shadow-[0_16px_40px_-12px_rgba(33,32,28,0.12)] ring-1 ring-sand-6/70">
          <img
            src="/screenshot.png"
            alt="Ech0 interface preview"
            className="block w-full align-middle"
            width={1412}
            height={1131}
            sizes="(max-width: 480px) 100vw, min(100%, 30rem)"
            decoding="sync"
            fetchPriority="high"
            loading="eager"
          />
        </div>
      </div>

      <main className="mx-auto w-full max-w-[min(100%,30rem)] px-5 pb-28 pt-8">
        {/* Hero: headline → positioning → CTA */}
        <section className="flex flex-col items-center gap-8 text-center">
          <h1 className="max-w-[22ch] font-serif text-[1.65rem] font-normal leading-[1.2] tracking-[-0.02em] text-sand-12 sm:text-[1.75rem]">
            After capture comes publishing.
          </h1>

          <div className="font-serif text-[0.9375rem] font-normal leading-[1.45] tracking-[0.01em] text-sand-11 [&>p+p]:mt-1">
            <p>Capture holds thoughts.</p>
            <p>Ech0 is what comes next—share and connect.</p>
            <p className="text-sand-12">
              Your timeline on your server.
            </p>
          </div>

          <div className="flex flex-col items-center gap-3">
            <Link
              to="/docs"
              prefetch="viewport"
              className="inline-flex items-center gap-2 rounded-full border border-sand-6 bg-sand-2/80 px-5 py-2.5 text-[13px] font-medium text-sand-11 shadow-[0_1px_2px_rgba(33,32,28,0.05)] no-underline transition-colors hover:border-sand-11/25 hover:bg-sand-2 hover:text-sand-12"
            >
              <LeafIcon className="opacity-75" />
              Get started
            </Link>
            <p className="text-[0.6875rem] font-medium tracking-wide text-sand-11/85">
              AGPL-3.0 · lightweight · open source
            </p>
          </div>
        </section>

        {/* Quote + philosophy */}
        <section className={`${dashedSection} text-left`}>
          <p className="font-serif text-base italic leading-snug text-sand-11">
            &ldquo;One timeline, purely.&rdquo;
          </p>
          <p className="mt-5 max-w-[34ch] text-[0.9375rem] leading-relaxed text-sand-10">
            Markdown, links, media—one stream, your rules.
          </p>
        </section>

        {/* Why choose */}
        <section className={dashedSection}>
          <h2 className="font-serif text-[1.125rem] font-semibold leading-snug text-sand-12">
            Why choose Ech0?
          </h2>
          <ol className="mt-6 list-decimal space-y-3 pl-[1.25rem] text-[0.9375rem] leading-relaxed text-sand-11 marker:font-serif marker:text-sand-11">
            <li>
              <span className="font-semibold text-sand-12">
                After capture
              </span>
              {" — "}
              Jotting is step one; Ech0 is publish and connect—not a private
              vault only.
            </li>
            <li>
              <span className="font-semibold text-sand-12">
                A real timeline
              </span>
              {" — "}
              A stream others can find—still your space, not a generic feed.
            </li>
            <li>
              <span className="font-semibold text-sand-12">Keep it quiet</span>
              {" — "}
              Self-hosted: your rules, no wiki sprawl or noise.
            </li>
          </ol>
        </section>

        {/* Why use */}
        <section className={dashedSection}>
          <h2 className="font-serif text-[1.125rem] font-semibold leading-snug text-sand-12">
            Why use Ech0?
          </h2>
          <ol className="mt-6 list-decimal space-y-3 pl-[1.25rem] text-[0.9375rem] leading-relaxed text-sand-11 marker:font-serif marker:text-sand-11">
            <li>
              <span className="font-semibold text-sand-12">
                Light &amp; self-hosted
              </span>
              {" — "}
              Notes, links, media—one timeline on your hardware.
            </li>
            <li>
              <span className="font-semibold text-sand-12">No lock-in</span>
              {" — "}
              Your server, your content—no ads or lock-in.
            </li>
            <li>
              <span className="font-semibold text-sand-12">Open &amp; easy</span>
              {" — "}
              AGPL-3.0; easy deploy. RSS, comments, multi-instance—connection
              without the noise.
            </li>
          </ol>
        </section>

        <footer className="mt-14 pt-12 text-center text-sm text-sand-11">
          <Link
            to="/docs"
            prefetch="viewport"
            className="font-medium underline-offset-4 transition-colors hover:text-sand-12"
          >
            Docs
          </Link>
          <span className="mx-2 text-sand-6">·</span>
          <Link
            to="/privacy"
            prefetch="viewport"
            className="font-medium underline-offset-4 transition-colors hover:text-sand-12"
          >
            Privacy
          </Link>
          <span className="mx-2 text-sand-6">·</span>
          <a
            href="https://github.com/lin-snow/Ech0"
            className="font-medium underline-offset-4 transition-colors hover:text-sand-12"
          >
            GitHub
          </a>
        </footer>
      </main>

      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(buildHomeJsonLd()) }}
      />
    </div>
  );
}
