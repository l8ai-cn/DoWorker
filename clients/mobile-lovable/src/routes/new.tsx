import { createFileRoute } from "@tanstack/react-router";
import { NewTaskPage } from "@/components/new-task/new-task-page";
import type { NewTaskSearch } from "@/components/new-task/use-new-task-state";
import { pageTitle } from "@/lib/app-brand";

export const Route = createFileRoute("/new")({
  validateSearch: (search: Record<string, unknown>): NewTaskSearch => ({
    expert: typeof search.expert === "string" ? search.expert : undefined,
    project: typeof search.project === "string" ? search.project : undefined,
    prompt: typeof search.prompt === "string" ? search.prompt : undefined,
  }),
  head: () => ({ meta: [{ title: pageTitle("下发新任务") }] }),
  component: NewTaskRoute,
});

function NewTaskRoute() {
  return <NewTaskPage search={Route.useSearch()} />;
}
