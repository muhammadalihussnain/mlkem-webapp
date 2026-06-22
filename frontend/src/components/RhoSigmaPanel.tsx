import React from 'react';
import { RhoSigmaPayload } from '../types/protocol';

interface RhoSigmaPanelProps {
  rhoSigma: RhoSigmaPayload | null;
}

const FIELDS: Array<{
  key: keyof RhoSigmaPayload;
  name: string;
  hint: string;
  tooltip: string;
  color: string;
}> = [
  {
    key: 'seed',
    name: 'seed',
    hint: 'OS / HW entropy',
    tooltip:
      'Random seed (32 bytes) drawn from the OS CSPRNG (e.g. /dev/urandom) or a hardware entropy source. ' +
      'It is hashed via G(seed) → (ρ, σ) to derive the public and private randomness.',
    color: '#f59e0b',
  },
  {
    key: 'rho',
    name: 'ρ',
    hint: 'public nonce',
    tooltip:
      'Public seed ρ (32 bytes). Deterministically expands into the public matrix A via SHAKE-128 (XOF). ' +
      'ρ is included in the public key — anyone can regenerate A from it.',
    color: '#60a5fa',
  },
  {
    key: 'sigma',
    name: 'σ',
    hint: 'secret nonce',
    tooltip:
      'Secret seed σ (32 bytes). Used with a PRF to sample the secret vector s and error vector e from ' +
      'the centered binomial distribution Bη. σ must remain private.',
    color: '#a78bfa',
  },
];

/**
 * RhoSigmaPanel shows the seed, ρ, and σ hex values each in a labelled
 * field with a hover tooltip explaining what each value is.
 */
export function RhoSigmaPanel({ rhoSigma }: RhoSigmaPanelProps) {
  if (!rhoSigma) {
    return (
      <p className="rho-sigma-panel--empty" style={{ color: 'var(--text-muted)', fontSize: 13, textAlign: 'center', padding: '20px 0', fontStyle: 'italic' }}>
        Click "Generate ρ / σ" to derive randomness from the OS entropy source.
      </p>
    );
  }

  return (
    <div className="rho-sigma-panel">
      {FIELDS.map(({ key, name, hint, tooltip, color }) => (
        <div className="rho-sigma-field" key={key}>
          <div
            className="rho-sigma-field__label"
            title={tooltip}
            aria-label={`${name}: ${hint}. ${tooltip}`}
            role="tooltip"
          >
            <span className="rho-sigma-field__label-name" style={{ color }}>
              {name}
            </span>
            <span className="rho-sigma-field__label-hint">{hint}</span>
          </div>
          <span className="rho-sigma-field__value" title={rhoSigma[key]}>
            {rhoSigma[key]}
          </span>
        </div>
      ))}
    </div>
  );
}
