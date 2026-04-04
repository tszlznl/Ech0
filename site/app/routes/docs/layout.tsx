import { Outlet } from "react-router";
import type { Route } from "./+types/layout";

export function meta({}: Route.MetaArgs) {
  return [{ title: "Documentation — Ech0" }];
}

export default function DocsLayout() {
  return (
    <div className="min-h-screen bg-app">
      <div className="mx-auto w-full max-w-[min(100%,30rem)] px-5 pb-24">
        <Outlet />
      </div>
    </div>
  );
}
