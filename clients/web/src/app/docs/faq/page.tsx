"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { FaqItem } from "@/components/docs/FaqItem";
import { FAQ_SECTIONS } from "./faq-sections";

export default function FAQPage() {
  const t = useTranslations();

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: FAQ_SECTIONS.flatMap((section) =>
      section.items.map(([questionKey, answerKey]) => ({
        "@type": "Question",
        name: t(questionKey),
        acceptedAnswer: {
          "@type": "Answer",
          text: t(answerKey),
        },
      })),
    ),
  };

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">{t("docs.faq.title")}</h1>
      <p className="text-muted-foreground leading-relaxed mb-8">{t("docs.faq.description")}</p>

      {FAQ_SECTIONS.map((section) => (
        <section key={section.categoryKey} className="mb-12">
          <h2 className="text-2xl font-semibold mb-4 text-foreground">{t(section.categoryKey)}</h2>
          <div className="space-y-3">
            {section.items.map(([questionKey, answerKey]) => (
              <FaqItem key={questionKey} questionKey={questionKey} answerKey={answerKey} />
            ))}
          </div>
        </section>
      ))}

      <DocNavigation />

      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
    </div>
  );
}
