import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { StepButton } from './StepButton';

describe('StepButton', () => {
  it('shows the waiting icon when step is not yet active', () => {
    render(
      <StepButton label="Generate A" onClick={vi.fn()} disabled={false} stepNumber={2} currentStep={0} />
    );
    expect(screen.getByRole('button')).toBeDisabled();
    expect(screen.getByText(/⏳/)).toBeInTheDocument();
  });

  it('shows the active icon when it is the next step', () => {
    render(
      <StepButton label="Generate A" onClick={vi.fn()} disabled={false} stepNumber={2} currentStep={1} />
    );
    expect(screen.getByText(/▶️/)).toBeInTheDocument();
    expect(screen.getByRole('button')).not.toBeDisabled();
  });

  it('shows the completed icon when step is done', () => {
    render(
      <StepButton label="Generate A" onClick={vi.fn()} disabled={false} stepNumber={2} currentStep={3} />
    );
    expect(screen.getByText(/✅/)).toBeInTheDocument();
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('calls onClick when the active step button is clicked', () => {
    const onClick = vi.fn();
    render(
      <StepButton label="Generate A" onClick={onClick} disabled={false} stepNumber={1} currentStep={0} />
    );
    fireEvent.click(screen.getByRole('button'));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('is disabled by the master disabled flag even when active', () => {
    render(
      <StepButton label="Generate A" onClick={vi.fn()} disabled={true} stepNumber={1} currentStep={0} />
    );
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('renders the label text', () => {
    render(
      <StepButton label="Compute t" onClick={vi.fn()} disabled={false} stepNumber={1} currentStep={0} />
    );
    expect(screen.getByText(/Compute t/)).toBeInTheDocument();
  });
});
