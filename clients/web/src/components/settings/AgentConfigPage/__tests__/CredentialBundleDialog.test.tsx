import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { CredentialBundleDialog } from "../CredentialBundleDialog";

const t = (key: string) => {
  const messages: Record<string, string> = {
    "common.cancel": "Cancel",
    "common.create": "Create",
    "common.saving": "Saving",
    "settings.agentConfig.credentialBundles.addTitle": "Add credentials",
    "settings.agentConfig.credentialBundles.description": "Worker credentials",
    "settings.agentConfig.credentialBundles.name": "Name",
    "settings.agentConfig.credentialBundles.namePlaceholder": "Personal Cursor key",
    "settings.agentConfig.credentialBundles.descriptionLabel": "Description",
    "settings.credentialForm.cursor.apiKey": "Cursor API Key",
  };
  return messages[key] ?? key;
};

describe("CredentialBundleDialog", () => {
  it("submits API-derived Cursor credential fields as an encrypted bundle payload", async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined);

    render(
      <CredentialBundleDialog
        open
        onOpenChange={vi.fn()}
        agentSlug="cursor-cli"
        credentialFields={[{ name: "CURSOR_API_KEY", type: "secret", optional: true }]}
        editing={null}
        onSubmit={onSubmit}
        t={t}
      />,
    );

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "cursor-main" },
    });
    fireEvent.change(document.getElementById("cred-CURSOR_API_KEY")!, {
      target: { value: "cursor-secret" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        {
          name: "cursor-main",
          description: "",
          data: { CURSOR_API_KEY: "cursor-secret" },
        },
        null,
      );
    });
  });
});
