import type { CompileResult } from "../domain/loop-types";
import type { OutputTab } from "../persistence/workspace-storage";
import type { SimulationEvidence } from "../simulation/run-simulation";

interface OutputPanelProps {
  compileResult: CompileResult;
  evidence: SimulationEvidence[];
  onFocusBlock: (blockId: string) => void;
  onTabChange: (tab: OutputTab) => void;
  tab: OutputTab;
}

const TABS: { id: OutputTab; label: string }[] = [
  { id: "diagnostics", label: "诊断" },
  { id: "json", label: "Loop JSON" },
  { id: "evidence", label: "运行证据" },
];

export function OutputPanel({
  compileResult,
  evidence,
  onFocusBlock,
  onTabChange,
  tab,
}: OutputPanelProps) {
  return (
    <section className="output-panel">
      <div className="output-tabs" role="tablist">
        {TABS.map(({ id, label }) => (
          <button
            aria-selected={tab === id}
            className={tab === id ? "active" : ""}
            key={id}
            onClick={() => onTabChange(id)}
            role="tab"
            type="button"
          >
            {label}
            {id === "diagnostics" && compileResult.diagnostics.length > 0 && (
              <span>{compileResult.diagnostics.length}</span>
            )}
          </button>
        ))}
      </div>
      <div className="output-content">
        {tab === "diagnostics" && (
          compileResult.diagnostics.length === 0 ? (
            <div className="output-empty">结构校验通过</div>
          ) : (
            <div className="diagnostic-list">
              {compileResult.diagnostics.map((diagnostic, index) => (
                <button
                  key={`${diagnostic.code}-${index}`}
                  onClick={() => diagnostic.blockId &&
                    onFocusBlock(diagnostic.blockId)}
                  type="button"
                >
                  <code>{diagnostic.code}</code>
                  <span>{diagnostic.message}</span>
                </button>
              ))}
            </div>
          )
        )}
        {tab === "json" && (
          compileResult.program ? (
            <pre>{JSON.stringify(compileResult.program, null, 2)}</pre>
          ) : (
            <div className="output-empty">修复诊断后才能生成 Loop JSON</div>
          )
        )}
        {tab === "evidence" && (
          evidence.length === 0 ? (
            <div className="output-empty">暂无运行证据</div>
          ) : (
            <div className="evidence-list">
              {evidence.map((event) => (
                <button
                  key={event.id}
                  onClick={() => onFocusBlock(event.blockId)}
                  type="button"
                >
                  <time>{new Date(event.timestamp).toLocaleTimeString()}</time>
                  <strong>{event.kind}</strong>
                  <span>{event.message}</span>
                </button>
              ))}
            </div>
          )
        )}
      </div>
    </section>
  );
}
