import { useState } from "react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { render, screen } from "@/test/test-utils";
import { createWorkerTemplateDraft } from "./worker-template-draft";
import {
  ResourceEditorSessionProvider,
  useResourceEditorSession,
} from "./ResourceEditorSessionProvider";
import { createResourceDraftState } from "./resource-draft-reducer";

describe("ResourceEditorSessionProvider", () => {
  it("keeps a WorkerTemplate draft when its editor remounts", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <SessionHarness mounted />,
    );

    await user.click(screen.getByRole("button", { name: "Select DoAgent" }));
    expect(screen.getByTestId("worker-type")).toHaveTextContent("do-agent");

    rerender(<SessionHarness mounted={false} />);
    rerender(<SessionHarness mounted />);

    expect(screen.getByTestId("worker-type")).toHaveTextContent("do-agent");
  });
});

function SessionHarness({ mounted }: { mounted: boolean }) {
  return (
    <ResourceEditorSessionProvider>
      {mounted && <DraftProbe />}
    </ResourceEditorSessionProvider>
  );
}

function DraftProbe() {
  const [initialState] = useState(() =>
    createResourceDraftState(createWorkerTemplateDraft("acme")),
  );
  const session = useResourceEditorSession(
    "worker-template:acme",
    initialState,
  );
  if (!session) throw new Error("Expected resource editor session.");
  const draft = session.state.draft;

  return (
    <>
      <output data-testid="worker-type">{draft.spec.workerType}</output>
      <button
        type="button"
        onClick={() => session.dispatch({
          type: "replace_draft",
          draft: {
            ...draft,
            spec: { ...draft.spec, workerType: "do-agent" },
          },
        })}
      >
        Select DoAgent
      </button>
    </>
  );
}
