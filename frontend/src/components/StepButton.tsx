import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';

interface StepButtonProps {
  label: string;
  onClick: () => void;
  disabled: boolean;
  stepNumber: number;
  currentStep: number;
  tooltip: string;
}

interface TooltipPos { top: number; left: number; width: number; }

export function StepButton({
  label, onClick, disabled, stepNumber, currentStep, tooltip,
}: StepButtonProps) {
  const completed = currentStep >= stepNumber;
  const active    = currentStep === stepNumber - 1;

  const [pos, setPos] = useState<TooltipPos | null>(null);
  const btnRef = useRef<HTMLButtonElement>(null);

  const show = () => {
    if (!btnRef.current) return;
    const r = btnRef.current.getBoundingClientRect();
    setPos({ top: r.top + window.scrollY, left: r.left + window.scrollX, width: r.width });
  };

  const hide = () => setPos(null);

  // clean up if button unmounts while hovered
  useEffect(() => () => setPos(null), []);

  let icon = '○';
  if (completed) icon = '✓';
  else if (active) icon = '▶';

  return (
    <>
      <button
        ref={btnRef}
        type="button"
        className={[
          'step-button',
          completed ? 'step-button--completed' : '',
          active     ? 'step-button--active'    : '',
        ].filter(Boolean).join(' ')}
        onClick={onClick}
        disabled={disabled || (!active && !completed)}
        aria-label={`${label} — step ${stepNumber}`}
        onMouseEnter={show}
        onMouseLeave={hide}
        onFocus={show}
        onBlur={hide}
      >
        <span className="step-button__icon" aria-hidden="true">{icon}</span>
        <span className="step-button__label">{label}</span>
        {/* Hidden span keeps tooltip text in DOM for tests and screen readers */}
        {tooltip && (
          <span className="step-button__tooltip-sr" aria-hidden="true">{tooltip}</span>
        )}
      </button>

      {pos && tooltip && createPortal(
        <div
          className="step-button__tooltip-portal"
          style={{
            position: 'absolute',
            top:  pos.top - 8,
            left: pos.left + pos.width / 2,
            transform: 'translate(-50%, -100%)',
            zIndex: 9999,
            pointerEvents: 'none',
          }}
          role="tooltip"
        >
          {tooltip}
          <span className="step-button__tooltip-portal-caret" aria-hidden="true" />
        </div>,
        document.body,
      )}
    </>
  );
}
