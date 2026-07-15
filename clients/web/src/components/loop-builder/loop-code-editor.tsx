"use client";

import { basicSetup } from "codemirror";
import { Compartment, EditorState } from "@codemirror/state";
import { EditorView } from "@codemirror/view";
import { useEffect, useRef } from "react";

interface LoopCodeEditorProps {
  value: string;
  readOnly: boolean;
  onChange: (value: string) => void;
}

export function LoopCodeEditor({ value, readOnly, onChange }: LoopCodeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | undefined>(undefined);
  const readOnlyCompartment = useRef(new Compartment());
  const onChangeRef = useRef(onChange);
  const syncingRef = useRef(false);

  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  useEffect(() => {
    if (!containerRef.current) return;
    const view = new EditorView({
      parent: containerRef.current,
      state: EditorState.create({
        doc: value,
        extensions: [
          basicSetup,
          EditorView.lineWrapping,
          EditorView.theme({
            "&": {
              height: "100%",
              backgroundColor: "var(--surface-raised)",
              color: "var(--foreground)",
              fontSize: "13px",
            },
            ".cm-content": {
              fontFamily: "var(--font-geist-mono), ui-monospace, monospace",
              padding: "14px 0",
            },
            ".cm-gutters": {
              backgroundColor: "var(--surface)",
              borderRight: "1px solid var(--border)",
              color: "var(--muted-foreground)",
            },
            ".cm-activeLine, .cm-activeLineGutter": {
              backgroundColor: "var(--surface-muted)",
            },
            ".cm-cursor": { borderLeftColor: "var(--primary)" },
          }),
          readOnlyCompartment.current.of(EditorState.readOnly.of(readOnly)),
          EditorView.updateListener.of((update) => {
            if (update.docChanged && !syncingRef.current) {
              onChangeRef.current(update.state.doc.toString());
            }
          }),
        ],
      }),
    });
    viewRef.current = view;
    return () => {
      view.destroy();
      viewRef.current = undefined;
    };
    // The editor is constructed once; later value/readOnly updates use transactions.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const view = viewRef.current;
    if (!view || view.state.doc.toString() === value) return;
    syncingRef.current = true;
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: value },
    });
    syncingRef.current = false;
  }, [value]);

  useEffect(() => {
    viewRef.current?.dispatch({
      effects: readOnlyCompartment.current.reconfigure(
        EditorState.readOnly.of(readOnly),
      ),
    });
  }, [readOnly]);

  return <div className="h-full min-h-0 overflow-hidden" ref={containerRef} />;
}
