export const FAQ_SECTIONS = [
  {
    categoryKey: "docs.faq.categories.runner",
    items: [
      ["docs.faq.items.runnerConnection.question", "docs.faq.items.runnerConnection.answer"],
      ["docs.faq.items.runnerMultiple.question", "docs.faq.items.runnerMultiple.answer"],
    ],
  },
  {
    categoryKey: "docs.faq.categories.pod",
    items: [
      ["docs.faq.items.podCreationFail.question", "docs.faq.items.podCreationFail.answer"],
      ["docs.faq.items.podStuck.question", "docs.faq.items.podStuck.answer"],
    ],
  },
  {
    categoryKey: "docs.faq.categories.apiKey",
    items: [
      ["docs.faq.items.apiKeyFormat.question", "docs.faq.items.apiKeyFormat.answer"],
      ["docs.faq.items.apiKeyMultiple.question", "docs.faq.items.apiKeyMultiple.answer"],
    ],
  },
  {
    categoryKey: "docs.faq.categories.git",
    items: [
      ["docs.faq.items.gitCloneFail.question", "docs.faq.items.gitCloneFail.answer"],
      ["docs.faq.items.gitWorktreeConflict.question", "docs.faq.items.gitWorktreeConflict.answer"],
    ],
  },
  {
    categoryKey: "docs.faq.categories.billing",
    items: [
      ["docs.faq.items.billingBYOK.question", "docs.faq.items.billingBYOK.answer"],
      ["docs.faq.items.billingFree.question", "docs.faq.items.billingFree.answer"],
    ],
  },
] as const;
