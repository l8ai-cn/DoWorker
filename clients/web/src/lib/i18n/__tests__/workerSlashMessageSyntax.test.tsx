import { render, screen } from "@testing-library/react";
import { NextIntlClientProvider, useTranslations } from "next-intl";
import { describe, expect, it } from "vitest";
import enMessages from "@/messages/en/app.json";
import zhMessages from "@/messages/zh/app.json";

function ModelCommand() {
  return <span>{useTranslations("workerSlash")("commands.model")}</span>;
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
