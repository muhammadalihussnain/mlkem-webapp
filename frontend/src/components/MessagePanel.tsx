import React from 'react';
import { EncryptResultPayload, DecryptResultPayload } from '../types/protocol';

/** Number of hex characters to preview for ciphertext. */
const CT_PREVIEW_LEN = 48;

interface MessagePanelProps {
  /** Called when the user clicks "Encrypt & Exchange". */
  onSendMessage: () => void;
  encryptResult: EncryptResultPayload | null;
  decryptResult: DecryptResultPayload | null;
  /** Disabled when keys are not yet available or not connected. */
  disabled: boolean;
}

/**
 * MessagePanel triggers the encapsulate → decapsulate round-trip and displays
 * the resulting ciphertext, shared secrets, and whether they match.
 *
 * Note: ML-KEM does not encrypt arbitrary plaintext — it encapsulates a random
 * shared secret.  The "message" field shown here is the random value embedded
 * in the ciphertext, not user-supplied text.
 */
export function MessagePanel({
  onSendMessage,
  encryptResult,
  decryptResult,
  disabled,
}: MessagePanelProps) {
  return (
    <div className="message-panel" aria-label="key exchange panel">
      <button
        type="button"
        className="message-panel__btn"
        onClick={onSendMessage}
        disabled={disabled}
        aria-label="run encapsulation and decapsulation"
      >
        🔐 Encapsulate &amp; Decapsulate
      </button>

      {encryptResult && (
        <section className="message-panel__result" aria-label="encapsulation result">
          <h4>Encapsulation</h4>
          <dl>
            <dt>Ciphertext ({encryptResult.ciphertext_size} B)</dt>
            <dd className="message-panel__hex" title={encryptResult.ciphertext}>
              {encryptResult.ciphertext.slice(0, CT_PREVIEW_LEN)}…
            </dd>
            <dt>Shared secret (encapsulator)</dt>
            <dd className="message-panel__hex">{encryptResult.shared_secret}</dd>
          </dl>
        </section>
      )}

      {decryptResult && (
        <section
          className={`message-panel__result message-panel__result--${decryptResult.match ? 'ok' : 'fail'}`}
          aria-label="decapsulation result"
        >
          <h4>Decapsulation</h4>
          <dl>
            <dt>Shared secret (decapsulator)</dt>
            <dd className="message-panel__hex">{decryptResult.shared_secret}</dd>
            <dt>Match</dt>
            <dd aria-label={`secrets match: ${decryptResult.match}`}>
              {decryptResult.match ? '✅ Secrets match' : '❌ Mismatch'}
            </dd>
          </dl>
        </section>
      )}
    </div>
  );
}
