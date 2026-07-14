import { X } from "lucide-react";
import { useState } from "react";

import {
  createCustomBlockDefinition,
  type CustomBlockDefinition,
} from "../custom-blocks/custom-block-definition";

interface CustomBlockDialogProps {
  open: boolean;
  onClose: () => void;
  onCreate: (definition: CustomBlockDefinition) => void;
}

export function CustomBlockDialog({
  open,
  onClose,
  onCreate,
}: CustomBlockDialogProps) {
  const [name, setName] = useState("");
  const [template, setTemplate] = useState("");
  const [errors, setErrors] = useState<string[]>([]);
  if (!open) return null;

  const submit = () => {
    const result = createCustomBlockDefinition({
      id: `macro-${Date.now().toString(36)}`,
      name,
      template,
    });
    if (!result.definition) {
      setErrors(result.errors);
      return;
    }
    onCreate(result.definition);
    setName("");
    setTemplate("");
    setErrors([]);
    onClose();
  };

  return (
    <div className="dialog-backdrop" role="presentation">
      <section
        aria-labelledby="custom-block-title"
        aria-modal="true"
        className="custom-dialog"
        role="dialog"
      >
        <header>
          <div>
            <span className="eyebrow">声明式任务宏</span>
            <h2 id="custom-block-title">创建自定义积木</h2>
          </div>
          <button
            className="icon-button"
            onClick={onClose}
            title="关闭"
            type="button"
          >
            <X size={18} />
          </button>
        </header>
        <label className="field-control">
          <span>名称</span>
          <input
            autoFocus
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
        </label>
        <label className="field-control">
          <span>任务模板</span>
          <textarea
            rows={6}
            value={template}
            onChange={(event) => setTemplate(event.target.value)}
            placeholder="修复 {{file-path}} 并运行 {{test-command}}"
          />
        </label>
        {errors.length > 0 && (
          <div className="form-errors">
            {errors.map((error) => <div key={error}>{error}</div>)}
          </div>
        )}
        <footer>
          <button className="secondary-button" onClick={onClose} type="button">
            取消
          </button>
          <button className="primary-button" onClick={submit} type="button">
            创建积木
          </button>
        </footer>
      </section>
    </div>
  );
}
