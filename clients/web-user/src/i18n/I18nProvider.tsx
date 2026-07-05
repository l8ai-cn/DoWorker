import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { detectLocale, persistLocale, type Locale } from "./locale";
import { en, type MessageTree } from "./messages/en";
import { zh } from "./messages/zh";

const catalogs: Record<Locale, MessageTree> = { en, zh };

interface I18nContextValue {
  locale: Locale;
  t: MessageTree;
  setLocale: (locale: Locale) => void;
}

const I18nContext = createContext<I18nContextValue | null>(null);

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(detectLocale);
  const value = useMemo<I18nContextValue>(
    () => ({
      locale,
      t: catalogs[locale],
      setLocale: (next) => {
        persistLocale(next);
        setLocaleState(next);
      },
    }),
    [locale],
  );
  useEffect(() => {
    document.documentElement.lang = locale === "zh" ? "zh-CN" : "en";
  }, [locale]);
  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n(): I18nContextValue {
  const ctx = useContext(I18nContext);
  if (!ctx) {
    return { locale: "en", t: en, setLocale: () => {} };
  }
  return ctx;
}
