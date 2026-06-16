import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { KeyPanel } from './KeyPanel';
import { PublicKeyPayload } from '../types/protocol';

const mockPK: PublicKeyPayload = {
  public_key: 'deadbeef'.repeat(100), // 800 hex chars
  public_key_size: 800,
};

describe('KeyPanel', () => {
  it('shows placeholder when publicKey is null', () => {
    render(<KeyPanel publicKey={null} />);
    expect(screen.getByLabelText(/public key not available/i)).toBeInTheDocument();
  });

  it('displays the key size', () => {
    render(<KeyPanel publicKey={mockPK} />);
    expect(screen.getByText('800 bytes')).toBeInTheDocument();
  });

  it('shows a truncated hex preview', () => {
    render(<KeyPanel publicKey={mockPK} />);
    // Full key is 800 chars; preview is 48 + '…'
    expect(screen.getByText(/deadbeef.*…/)).toBeInTheDocument();
  });

  it('renders a copy button', () => {
    render(<KeyPanel publicKey={mockPK} />);
    expect(screen.getByLabelText(/copy public key/i)).toBeInTheDocument();
  });

  it('calls clipboard.writeText when copy is clicked', () => {
    const writeMock = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: writeMock },
      writable: true,
      configurable: true,
    });

    render(<KeyPanel publicKey={mockPK} />);
    fireEvent.click(screen.getByLabelText(/copy public key/i));
    expect(writeMock).toHaveBeenCalledWith(mockPK.public_key);
  });

  it('has accessible panel label', () => {
    render(<KeyPanel publicKey={mockPK} />);
    expect(screen.getByLabelText(/public key panel/i)).toBeInTheDocument();
  });
});
