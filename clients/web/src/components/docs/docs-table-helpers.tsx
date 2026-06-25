import type { ReactNode } from "react";
import type { DocsTableRow } from "./DocsTable";

export function docsLabel(text: string): ReactNode {
  return <span className="font-medium">{text}</span>;
}

export function docsMono(text: string): ReactNode {
  return <span className="font-mono text-xs">{text}</span>;
}

export function buildDocsRows(
  t: (key: string) => string,
  prefix: string,
  keys: string[],
  options?: { params?: boolean },
): DocsTableRow[] {
  return keys.map((key) => ({
    cells: options?.params
      ? [
          docsLabel(t(`${prefix}.${key}`)),
          t(`${prefix}.${key}Desc`),
          docsMono(t(`${prefix}.${key}Params`)),
        ]
      : [docsLabel(t(`${prefix}.${key}`)), t(`${prefix}.${key}Desc`)],
  }));
}

export function twoColumnHeaders(
  t: (key: string) => string,
  prefix: string,
  firstKey: string,
  secondKey: string,
) {
  return [
    { header: t(`${prefix}.${firstKey}`) },
    { header: t(`${prefix}.${secondKey}`) },
  ];
}

export function buildTripleKeyRows(
  t: (key: string) => string,
  prefix: string,
  entries: Array<[string, string, string]>,
  options?: { monoFirst?: boolean },
): DocsTableRow[] {
  return entries.map(([first, second, third]) => ({
    cells: [
      options?.monoFirst
        ? docsMono(t(`${prefix}.${first}`))
        : docsLabel(t(`${prefix}.${first}`)),
      t(`${prefix}.${second}`),
      t(`${prefix}.${third}`),
    ],
  }));
}

export function threeColumnHeaders(
  t: (key: string) => string,
  prefix: string,
  toolKey: string,
  descKey: string,
  paramsKey: string,
) {
  return [
    { header: t(`${prefix}.${toolKey}`) },
    { header: t(`${prefix}.${descKey}`) },
    { header: t(`${prefix}.${paramsKey}`) },
  ];
}
