interface ConfirmationDialogProps {
  body: string;
  confirmLabel: string;
  onClose: () => void;
  onConfirm: () => void;
  open: boolean;
  title: string;
}

export function ConfirmationDialog({
  body,
  confirmLabel,
  onClose,
  onConfirm,
  open,
  title,
}: ConfirmationDialogProps) {
  if (!open) return null;
  return (
    <div className="dialog-backdrop" role="presentation">
      <section
        aria-labelledby="confirmation-title"
        aria-modal="true"
        className="confirmation-dialog"
        role="dialog"
      >
        <h2 id="confirmation-title">{title}</h2>
        <p>{body}</p>
        <footer>
          <button className="secondary-button" onClick={onClose} type="button">
            取消
          </button>
          <button
            className="danger-button"
            onClick={() => {
              onConfirm();
              onClose();
            }}
            type="button"
          >
            {confirmLabel}
          </button>
        </footer>
      </section>
    </div>
  );
}
