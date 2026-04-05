/**
 * Canonical site origin for absolute URLs (OG, Twitter, JSON-LD).
 * Override with `VITE_SITE_URL` when deploying.
 * If you use a custom domain, update `public/sitemap.xml` and `public/robots.txt`
 * `Sitemap:` URL to match.
 */
export function siteUrl(): string {
  const raw = import.meta.env.VITE_SITE_URL ?? "https://www.ech0.app";
  return raw.replace(/\/$/, "");
}

export function absoluteUrl(path: string): string {
  const base = siteUrl();
  const p = path.startsWith("/") ? path : `/${path}`;
  return `${base}${p}`;
}

export const SITE_NAME = "Ech0";

export const DEFAULT_DESCRIPTION =
  "Let your thoughts flow: a personal timeline on your server—self-hosted, AGPL-3.0, ad-free and platform-free.";
