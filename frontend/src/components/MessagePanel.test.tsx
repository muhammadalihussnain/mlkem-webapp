import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MessagePanel } from './MessagePanel';
import { EncryptResultPayload, DecryptResultPayload } from '../types/protocol';

const mockEnc: EncryptResultPayload = {
  ciphertext: 'ab'.repeat(400),
  ciphertext_size: 768,
  shared_secret: 'ff'.repeat(32),
  message: '00'.repeat(32),
};

const mockDecMatch: DecryptResultPayload = {
  shared_secret: 'ff'.repeat(32),
  match: true,
};

const mockDecFail: DecryptResultPayload = {
  shared_secret: '11'.repeat(32),
  match: false,
};

describe('MessagePanel', () => {
  it('renders the action button', () => {
    render(
      <MessagePanel onSendMessage={vi.fn()} encryptResult={null} decryptResult={null} disabled={false} />
    );
    expect(screen.getByLabelText(/run encapsulation/i)).toBeInTheDocument();
  });

  it('disables the button when disabled=true', () => {
    render(
      <MessagePanel onSendMessage={vi.fn()} encryptResult={null} decryptResult={null} disabled={true} />
    );
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('calls onSendMessage when button is clicked', () => {
    const onSend = vi.fn();
    render(
      <MessagePanel onSendMessage={onSend} encryptResult={null} decryptResult={null} disabled={false} />
    );
    fireEvent.click(screen.getByRole('button'));
    expect(onSend).toHaveBeenCalledOnce();
  });

  it('shows encapsulation result when encryptResult is set', () => {
    render(
      <MessagePanel onSendMessage={vi.fn()} encryptResult={mockEnc} decryptResult={null} disabled={false} />
    );
    expect(screen.getByLabelText(/encapsulation result/i)).toBeInTheDocument();
    expect(screen.getByText(/768.*B/)).toBeInTheDocument();
  });

  it('shows decapsulation result when decryptResult is set', () => {
    render(
      <MessagePanel onSendMessage={vi.fn()} encryptResult={mockEnc} decryptResult={mockDecMatch} disabled={false} />
    );
    expect(screen.getByLabelText(/decapsulation result/i)).toBeInTheDocument();
    expect(screen.getByText(/secrets match/i)).toBeInTheDocument();
  });

  it('shows mismatch indicator when match=false', () => {
    render(
      <MessagePanel onSendMessage={vi.fn()} encryptResult={mockEnc} decryptResult={mockDecFail} disabled={false} />
    );
    expect(screen.getByText(/mismatch/i)).toBeInTheDocument();
  });

  it('does not render result sections when results are null', () => {
    render(
      <MessagePanel onSendMessage={vi.fn()} encryptResult={null} decryptResult={null} disabled={false} />
    );
    expect(screen.queryByLabelText(/encapsulation result/i)).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/decapsulation result/i)).not.toBeInTheDocument();
  });
});
