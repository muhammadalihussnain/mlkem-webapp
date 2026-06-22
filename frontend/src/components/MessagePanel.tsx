import type { EncryptResultPayload, DecryptResultPayload } from '../types/protocol';

const GLITTER = ['✦', '✧', '✶', '✸', '✺', '❋', '✵', '✴'];
const rng = (seed: string, i: number) => GLITTER[(seed.charCodeAt(i % seed.length) + i) % GLITTER.length];

interface MessagePanelProps {
  onSendMessage: () => void;
  encryptResult: EncryptResultPayload | null;
  decryptResult: DecryptResultPayload | null;
  disabled: boolean;
}

export function MessagePanel({ onSendMessage, encryptResult, decryptResult, disabled }: MessagePanelProps) {
  return (
    <div className="message-panel" aria-label="key exchange panel">
      <button
        type="button"
        className="message-panel__btn"
        onClick={onSendMessage}
        disabled={disabled}
        aria-label="run encapsulation and decapsulation"
      >
        Encapsulate &amp; Decapsulate
      </button>

      {encryptResult && (
        <section className="message-panel__result" aria-label="encapsulation result">
          <h4>Encapsulation Result</h4>

          {/* Ciphertext with glitter display */}
          <div className="cipher-box" aria-label="ciphertext">
            <div className="cipher-box__header">
              <span className="cipher-box__title">Ciphertext</span>
              <span className="cipher-box__size">{encryptResult.ciphertext_size} bytes</span>
            </div>
            <div className="cipher-box__glitter" aria-hidden="true">
              {Array.from({ length: 12 }, (_, i) => (
                <span key={i} className="cipher-box__star"
                  style={{ animationDelay: `${i * 0.18}s`, left: `${(i * 8.3) % 100}%` }}>
                  {rng(encryptResult.ciphertext, i)}
                </span>
              ))}
            </div>
            <div className="cipher-box__hex" title={encryptResult.ciphertext}>
              {encryptResult.ciphertext.slice(0, 64)}…
            </div>
            <div className="cipher-box__footer">
              {rng(encryptResult.ciphertext, 0)} Encapsulated shared secret &middot; {encryptResult.ciphertext_size} bytes
            </div>
          </div>

          <div className="message-panel__secret">
            <span className="message-panel__secret-label">Shared secret (encapsulator)</span>
            <span className="message-panel__hex">{encryptResult.shared_secret}</span>
          </div>
        </section>
      )}

      {decryptResult && (
        <section
          className={`message-panel__result message-panel__result--${decryptResult.match ? 'ok' : 'fail'}`}
          aria-label="decapsulation result"
        >
          <h4>Decapsulation Result</h4>
          <div className="message-panel__secret">
            <span className="message-panel__secret-label">Shared secret (decapsulator)</span>
            <span className="message-panel__hex">{decryptResult.shared_secret}</span>
          </div>
          <div className={`message-panel__match message-panel__match--${decryptResult.match ? 'ok' : 'fail'}`}
            aria-label={`secrets match: ${decryptResult.match}`}>
            {decryptResult.match ? '✅ Secrets match — key exchange successful!' : '❌ Mismatch — something went wrong.'}
          </div>
        </section>
      )}
    </div>
  );
}
