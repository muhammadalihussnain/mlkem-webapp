import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { FlavorSelector } from './FlavorSelector';
import { FLAVORS } from '../types/protocol';

describe('FlavorSelector', () => {
  it('renders the label and select element', () => {
    render(<FlavorSelector onSelect={vi.fn()} currentFlavor="" disabled={false} />);
    expect(screen.getByLabelText(/security level/i)).toBeInTheDocument();
  });

  it('renders all flavor options', () => {
    render(<FlavorSelector onSelect={vi.fn()} currentFlavor="" disabled={false} />);
    FLAVORS.forEach(f => {
      expect(screen.getByRole('option', { name: f.label })).toBeInTheDocument();
    });
  });

  it('reflects currentFlavor as the selected value', () => {
    render(<FlavorSelector onSelect={vi.fn()} currentFlavor="768" disabled={false} />);
    const select = screen.getByRole('combobox') as HTMLSelectElement;
    expect(select.value).toBe('768');
  });

  it('calls onSelect with the chosen value when changed', () => {
    const onSelect = vi.fn();
    render(<FlavorSelector onSelect={onSelect} currentFlavor="" disabled={false} />);
    fireEvent.change(screen.getByRole('combobox'), { target: { value: '512' } });
    expect(onSelect).toHaveBeenCalledWith('512');
  });

  it('disables the select when disabled=true', () => {
    render(<FlavorSelector onSelect={vi.fn()} currentFlavor="" disabled={true} />);
    expect(screen.getByRole('combobox')).toBeDisabled();
  });

  it('enables the select when disabled=false', () => {
    render(<FlavorSelector onSelect={vi.fn()} currentFlavor="" disabled={false} />);
    expect(screen.getByRole('combobox')).not.toBeDisabled();
  });
});
