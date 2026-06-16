import React from 'react';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ParameterPanel } from './ParameterPanel';
import { ParamsPayload } from '../types/protocol';

const mockParams: ParamsPayload = {
  flavor: '768',
  n: 256,
  q: 3329,
  k: 3,
  eta1: 2,
  eta2: 2,
  du: 10,
  dv: 4,
  pk_size: 1184,
  sk_size: 1152,
  ct_size: 1088,
};

describe('ParameterPanel', () => {
  it('shows a placeholder when params is null', () => {
    render(<ParameterPanel params={null} />);
    expect(screen.getByLabelText(/parameter panel empty/i)).toBeInTheDocument();
  });

  it('renders a table when params are provided', () => {
    render(<ParameterPanel params={mockParams} />);
    expect(screen.getByRole('table')).toBeInTheDocument();
  });

  it('displays the ring degree n', () => {
    render(<ParameterPanel params={mockParams} />);
    expect(screen.getByText('256')).toBeInTheDocument();
  });

  it('displays the modulus q', () => {
    render(<ParameterPanel params={mockParams} />);
    expect(screen.getByText('3329')).toBeInTheDocument();
  });

  it('displays the module rank k', () => {
    render(<ParameterPanel params={mockParams} />);
    expect(screen.getByText('3')).toBeInTheDocument();
  });

  it('displays the public key size', () => {
    render(<ParameterPanel params={mockParams} />);
    expect(screen.getByText('1184')).toBeInTheDocument();
  });

  it('displays the ciphertext size', () => {
    render(<ParameterPanel params={mockParams} />);
    expect(screen.getByText('1088')).toBeInTheDocument();
  });
});
