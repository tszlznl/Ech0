/** Canonical site origin for absolute URLs (OG, Twitter, JSON-LD). Override with VITE_SITE_URL when deploying. */
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
  "Self-hosted microblog and timeline you fully own. Publish short posts, links, and media; optional comments. Lightweight, easy to deploy, and open source.";
