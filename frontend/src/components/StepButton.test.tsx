import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { StepButton } from './StepButton';

const defaultProps = {
  onClick: vi.fn(),
  disabled: false,
  tooltip: 'Test tooltip description',
};

describe('StepButton', () => {
  it('is disabled and shows waiting icon when step is not yet active', () => {
    render(<StepButton {...defaultProps} label="Generate A" stepNumber={2} currentStep={0} />);
    expect(screen.getByRole('button')).toBeDisabled();
    expect(screen.getByText('○')).toBeInTheDocument();
  });

  it('shows active icon and is enabled when it is the next step', () => {
    render(<StepButton {...defaultProps} label="Generate A" stepNumber={2} currentStep={1} />);
    expect(screen.getByText('▶')).toBeInTheDocument();
    expect(screen.getByRole('button')).not.toBeDisabled();
  });

  it('shows completed icon when step is done', () => {
    render(<StepButton {...defaultProps} label="Generate A" stepNumber={2} currentStep={3} />);
    expect(screen.getByText('✓')).toBeInTheDocument();
  });

  it('calls onClick when the active step button is clicked', () => {
    const onClick = vi.fn();
    render(<StepButton {...defaultProps} onClick={onClick} label="Generate A" stepNumber={1} currentStep={0} />);
    fireEvent.click(screen.getByRole('button'));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('is disabled by the master disabled flag even when active', () => {
    render(<StepButton {...defaultProps} disabled={true} label="Generate A" stepNumber={1} currentStep={0} />);
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('renders the label text', () => {
    render(<StepButton {...defaultProps} label="Compute t" stepNumber={1} currentStep={0} />);
    expect(screen.getByText(/Compute t/)).toBeInTheDocument();
  });

  it('renders the tooltip text', () => {
    render(<StepButton {...defaultProps} label="Generate A" stepNumber={1} currentStep={0} tooltip="Expand rho into matrix A" />);
    expect(screen.getByText(/Expand rho into matrix A/)).toBeInTheDocument();
  });
});
