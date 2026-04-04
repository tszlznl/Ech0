import { Link } from "react-router";
import type { Route } from "./+types/index";
import { listDocCards } from "../../docs/registry";
import { absoluteUrl } from "../../site";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Documentation — Ech0" },
    {
      name: "description",
      content:
        "Ech0 documentation: deployment, features, and policies.",
    },
  ];
}

export function loader() {
  return { cards: listDocCards() };
}

export const links: Route.LinksFunction = () => [
  { rel: "canonical", href: absoluteUrl("/docs") },
];

export default function DocsIndex({ loaderData }: Route.ComponentProps) {
  const { cards } = loaderData;

  return (
    <>
      <header className="flex items-center justify-between pt-8">
        <Link
          to="/"
          className="text-xs font-medium text-sand-11 transition-colors hover:text-sand-12"
        >
          ← Home
        </Link>
      </header>

      <h1 className="mt-8 font-serif text-xl font-normal leading-snug text-sand-12">
        Documentation
      </h1>
      <p className="mt-2 text-[0.8125rem] leading-relaxed text-sand-11">
        浏览下方文档卡片进入各篇说明。
      </p>

      <ul className="mt-8 grid grid-cols-2 gap-2.5">
        {cards.map((card) => (
          <li key={card.slug}>
            <Link
              to={`/docs/${card.slug}`}
              className="flex aspect-[5/3] flex-col justify-between rounded-xl border border-sand-6 bg-sand-2/50 p-2.5 text-left shadow-[0_1px_2px_rgba(33,32,28,0.04)] transition-colors hover:border-sand-11/25 hover:bg-sand-2"
            >
              <span className="font-serif text-xs font-semibold leading-snug text-sand-12 line-clamp-3">
                {card.title}
              </span>
              {card.description ? (
                <span className="mt-1 line-clamp-2 text-[0.6875rem] leading-snug text-sand-11">
                  {card.description}
                </span>
              ) : null}
            </Link>
          </li>
        ))}
      </ul>
    </>
  );
}
