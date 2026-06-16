import React from 'react';

interface StepButtonProps {
  /** Human-readable step label. */
  label: string;
  /** Called when the button is clicked. */
  onClick: () => void;
  /** Master disabled flag (e.g. not connected). */
  disabled: boolean;
  /** The 1-based index of this step in the sequence. */
  stepNumber: number;
  /** The current progress index (number of completed steps). */
  currentStep: number;
}

/**
 * StepButton renders a single key-generation step.
 * It is only clickable when it is the next step to execute.
 */
export function StepButton({ label, onClick, disabled, stepNumber, currentStep }: StepButtonProps) {
  const completed = currentStep >= stepNumber;
  const active    = currentStep === stepNumber - 1;

  let icon = '⏳';
  if (completed) icon = '✅';
  else if (active) icon = '▶️';

  return (
    <button
      type="button"
      className={[
        'step-button',
        completed ? 'step-button--completed' : '',
        active     ? 'step-button--active'    : '',
      ].join(' ').trim()}
      onClick={onClick}
      disabled={disabled || !active}
      aria-label={`${label} step ${stepNumber}`}
    >
      <span aria-hidden="true">{icon}</span> {label}
    </button>
  );
}
