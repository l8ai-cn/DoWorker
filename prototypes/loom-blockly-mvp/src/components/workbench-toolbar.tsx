import {
  Blocks,
  Braces,
  CheckCircle2,
  LayoutTemplate,
  Play,
  Plus,
  Save,
  Square,
} from "lucide-react";

interface WorkbenchToolbarProps {
  dirty: boolean;
  problemCount: number;
  running: boolean;
  valid: boolean;
  workspaceReady: boolean;
  onGenerate: () => void;
  onLoadExample: () => void;
  onOpenCustom: () => void;
  onRun: () => void;
  onSave: () => void;
  onStop: () => void;
  onValidate: () => void;
}

export function WorkbenchToolbar({
  dirty,
  problemCount,
  running,
  valid,
  workspaceReady,
  onGenerate,
  onLoadExample,
  onOpenCustom,
  onRun,
  onSave,
  onStop,
  onValidate,
}: WorkbenchToolbarProps) {
  return (
    <header className="workbench-toolbar">
      <div className="brand-block">
        <span className="brand-mark"><Blocks size={18} /></span>
        <div>
          <strong>Loom 工作台</strong>
          <span>Goal Loop / Blockly MVP</span>
        </div>
      </div>
      <div className="toolbar-status">
        <span className={valid ? "status-valid" : "status-invalid"}>
          {valid ? "结构有效" : `${problemCount} 个问题`}
        </span>
        <span>{dirty ? "未保存" : "已保存"}</span>
      </div>
      <div className="toolbar-actions">
        <button
          className="icon-button"
          disabled={!workspaceReady || running}
          onClick={onLoadExample}
          title="载入示例"
          type="button"
        >
          <LayoutTemplate size={17} />
        </button>
        <button
          className="icon-button"
          disabled={running}
          onClick={onOpenCustom}
          title="创建自定义积木"
          type="button"
        >
          <Plus size={17} />
        </button>
        <button
          className="secondary-button"
          disabled={!dirty || running}
          onClick={onSave}
          type="button"
        >
          <Save size={16} /> 保存
        </button>
        <button className="secondary-button" onClick={onValidate} type="button">
          <CheckCircle2 size={16} /> 验证
        </button>
        <button
          className="secondary-button"
          disabled={!valid}
          onClick={onGenerate}
          type="button"
        >
          <Braces size={16} /> 生成 Loop
        </button>
        {running ? (
          <button className="danger-button" onClick={onStop} type="button">
            <Square size={15} /> 停止
          </button>
        ) : (
          <button
            className="primary-button"
            disabled={!valid}
            onClick={onRun}
            type="button"
          >
            <Play size={16} /> 运行模拟
          </button>
        )}
      </div>
    </header>
  );
}
