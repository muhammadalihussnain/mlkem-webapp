import React from 'react';
import { PublicKeyPayload } from '../types/protocol';

/** Number of hex characters to show before truncating. */
const HEX_PREVIEW_LEN = 48;

interface KeyPanelProps {
  publicKey: PublicKeyPayload | null;
}

/**
 * KeyPanel displays the encoded public key with a truncated hex preview
 * and a copy-to-clipboard button.
 */
export function KeyPanel({ publicKey }: KeyPanelProps) {
  if (!publicKey) {
    return (
      <p className="key-panel--empty" aria-label="public key not available">
        Public key not yet generated.
      </p>
    );
  }

  const preview = publicKey.public_key.slice(0, HEX_PREVIEW_LEN) + '…';

  const handleCopy = () => {
    navigator.clipboard.writeText(publicKey.public_key).catch(console.error);
  };

  return (
    <div className="key-panel" aria-label="public key panel">
      <dl className="key-panel__fields">
        <dt>Size</dt>
        <dd>{publicKey.public_key_size} bytes</dd>

        <dt>Public key (hex)</dt>
        <dd className="key-panel__hex">
          <span title={publicKey.public_key}>{preview}</span>
          <button
            type="button"
            className="key-panel__copy"
            onClick={handleCopy}
            aria-label="copy public key"
          >
            📋
          </button>
        </dd>
      </dl>
    </div>
  );
}
