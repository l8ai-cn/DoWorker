import { render, screen } from "@testing-library/react";
import { NextIntlClientProvider, useTranslations } from "next-intl";
import { describe, expect, it } from "vitest";
import enMessages from "@/messages/en/app.json";
import enIdeMessages from "@/messages/en/ide.json";
import zhMessages from "@/messages/zh/app.json";

function ModelCommand() {
  return <span>{useTranslations("workerSlash")("commands.model")}</span>;
}

function RunnerHostHint() {
  return <span>{useTranslations("ide")("createPod.runnerHostHint")}</span>;
}

describe("worker slash command messages", () => {
  it.each([
    ["en", enMessages, "Switch model: /model [name] | default"],
    ["zh", zhMessages, "切换模型：/model [名称] | default"],
  ])("formats the model command in %s", (locale, messages, expected) => {
    render(
      <NextIntlClientProvider locale={locale} messages={messages}>
        <ModelCommand />
      </NextIntlClientProvider>,
    );

    expect(screen.getByText(expected)).toBeInTheDocument();
  });
});

describe("worker creation messages", () => {
  it("renders the cluster help text in English", () => {
    render(
      <NextIntlClientProvider locale="en" messages={enIdeMessages}>
        <RunnerHostHint />
      </NextIntlClientProvider>,
    );

    expect(
      screen.getByText(
        "Clusters provide compute. Workers are isolated workspaces started from the selected image.",
      ),
    ).toBeInTheDocument();
  });
});
