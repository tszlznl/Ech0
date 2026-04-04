import { Link, redirect } from "react-router";
import type { Route } from "./+types/doc";
import { getDoc } from "../../docs/registry";
import { MarkdownDoc } from "../../docs/MarkdownDoc";
import { absoluteUrl } from "../../site";

export function meta({ data }: Route.MetaArgs) {
  if (!data?.title) {
    return [{ title: "Documentation — Ech0" }];
  }
  return [
    { title: `${data.title} — Ech0 Docs` },
    {
      name: "description",
      content: data.description || `${data.title} — Ech0 documentation.`,
    },
    { property: "og:title", content: `${data.title} — Ech0 Docs` },
    { property: "og:url", content: absoluteUrl(`/docs/${data.slug}`) },
  ];
}

export function loader({ params }: Route.LoaderArgs) {
  const splat = params["*"]?.replace(/\/$/, "") ?? "";
  if (splat === "" || splat === "README") {
    return redirect("/docs");
  }
  const doc = getDoc(splat);
  if (!doc) throw new Response("Not Found", { status: 404 });
  return {
    body: doc.body,
    title: doc.title,
    description: doc.description,
    slug: splat,
  };
}

export default function DocsPage({ loaderData }: Route.ComponentProps) {
  const { body } = loaderData;

  return (
    <>
      <div className="flex items-center justify-between pt-8">
        <Link
          to="/docs"
          className="text-xs font-medium text-sand-11 transition-colors hover:text-sand-12"
        >
          ← All docs
        </Link>
        <Link
          to="/"
          className="text-xs font-medium text-sand-11 transition-colors hover:text-sand-12"
        >
          Home
        </Link>
      </div>

      <article className="mt-8">
        <MarkdownDoc content={body} />
      </article>
    </>
  );
}
