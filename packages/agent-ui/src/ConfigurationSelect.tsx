import { Check, ChevronDown, LoaderCircle } from "lucide-react";
import {
  useEffect,
  useId,
  useRef,
  useState,
  type KeyboardEvent,
} from "react";

import type { AgentConfigurationOption } from "./contracts";

export function ConfigurationSelect({
  disabled,
  label,
  onChange,
  options,
  optionsLabel,
  placeholder,
  value,
}: {
  disabled: boolean;
  label: string;
  onChange: (value: string) => void;
  options: AgentConfigurationOption[];
  optionsLabel: string;
  placeholder: string;
  value: string;
}) {
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const rootRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const optionRefs = useRef<Array<HTMLButtonElement | null>>([]);
  const listboxId = useId();
  const selectedIndex = options.findIndex((option) => option.value === value);
  const focusIndex = selectedIndex >= 0 ? selectedIndex : 0;
  const selected = selectedIndex >= 0 ? options[selectedIndex] : undefined;
  const hasChoice = options.some((option) => option.value !== value);

  useEffect(() => {
    if (!open) return;
    const close = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    };
    document.addEventListener("pointerdown", close);
    return () => document.removeEventListener("pointerdown", close);
  }, [open]);

  useEffect(() => {
    if (!open) return;
    setActiveIndex(focusIndex);
    requestAnimationFrame(() => optionRefs.current[focusIndex]?.focus());
  }, [focusIndex, open]);

  const move = (offset: number) => {
    const next = (activeIndex + offset + options.length) % options.length;
    setActiveIndex(next);
    optionRefs.current[next]?.focus();
  };

  const choose = (nextValue: string) => {
    setOpen(false);
    requestAnimationFrame(() => triggerRef.current?.focus());
    if (nextValue !== value) onChange(nextValue);
  };

  const handleOptionKeyDown = (
    event: KeyboardEvent<HTMLButtonElement>,
    index: number,
  ) => {
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      move(event.key === "ArrowDown" ? 1 : -1);
    } else if (event.key === "Home" || event.key === "End") {
      event.preventDefault();
      const next = event.key === "Home" ? 0 : options.length - 1;
      setActiveIndex(next);
      optionRefs.current[next]?.focus();
    } else if (event.key === "Escape") {
      event.preventDefault();
      setOpen(false);
      triggerRef.current?.focus();
    } else if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      choose(options[index].value);
    }
  };

  return (
    <div
      className="relative min-w-0"
      onBlur={(event) => {
        if (
          open &&
          !event.currentTarget.contains(event.relatedTarget as Node | null)
        ) {
          setOpen(false);
        }
      }}
      ref={rootRef}
    >
      <button
        aria-controls={listboxId}
        aria-expanded={open}
        aria-haspopup="listbox"
        aria-label={label}
        className="flex h-8 max-w-40 items-center gap-1.5 rounded-md px-2 text-xs text-foreground outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-default disabled:text-muted-foreground"
        disabled={disabled || !hasChoice}
        onClick={() => setOpen((current) => !current)}
        onKeyDown={(event) => {
          if (event.key !== "ArrowDown" && event.key !== "ArrowUp") return;
          event.preventDefault();
          setOpen(true);
        }}
        role="combobox"
        ref={triggerRef}
        title={label}
        type="button"
      >
        <span className="truncate">{selected?.label ?? placeholder}</span>
        {disabled ? (
          <LoaderCircle className="size-3.5 shrink-0 animate-spin" />
        ) : (
          <ChevronDown
            className={`size-3.5 shrink-0 transition-transform ${open ? "rotate-180" : ""}`}
          />
        )}
      </button>
      {open && (
        <div
          aria-label={optionsLabel}
          className="absolute bottom-full left-0 z-30 mb-2 w-max min-w-52 max-w-72 overflow-hidden rounded-md border border-border bg-popover p-1 shadow-lg"
          id={listboxId}
          role="listbox"
        >
          {options.map((option, index) => {
            const selectedOption = option.value === value;
            return (
              <button
                aria-selected={selectedOption}
                className="flex min-h-10 w-full items-start gap-2 rounded px-2.5 py-2 text-left text-sm outline-none hover:bg-muted focus:bg-muted"
                key={option.value}
                onClick={() => choose(option.value)}
                onFocus={() => setActiveIndex(index)}
                onKeyDown={(event) => handleOptionKeyDown(event, index)}
                ref={(element) => {
                  optionRefs.current[index] = element;
                }}
                role="option"
                tabIndex={index === activeIndex ? 0 : -1}
                type="button"
              >
                <Check
                  className={`mt-0.5 size-4 shrink-0 ${selectedOption ? "opacity-100" : "opacity-0"}`}
                />
                <span className="min-w-0">
                  <span className="block font-medium">{option.label}</span>
                  {option.description && (
                    <span className="mt-0.5 block text-xs text-muted-foreground">
                      {option.description}
                    </span>
                  )}
                </span>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
