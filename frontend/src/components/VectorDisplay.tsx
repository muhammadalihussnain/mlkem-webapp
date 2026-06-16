import React from 'react';
import { VectorsPayload, TComputedPayload } from '../types/protocol';

const PREVIEW_COEFFS = 8;

interface VectorDisplayProps {
  vectors: VectorsPayload | null;
  tComputed: TComputedPayload | null;
}

interface PolyRowProps {
  label: string;
  poly: number[];
}

function PolyRow({ label, poly }: PolyRowProps) {
  const preview = poly.slice(0, PREVIEW_COEFFS).join(', ');
  return (
    <div className="vector-display__row">
      <span className="vector-display__label">{label}</span>
      <span className="vector-display__coeffs" title={`Full polynomial ${label}`}>
        [{preview}…]
      </span>
    </div>
  );
}

/**
 * VectorDisplay renders the secret vector s, error vector e, and (once
 * computed) the public vector t, each as a collapsed polynomial row showing
 * the first PREVIEW_COEFFS coefficients.
 */
export function VectorDisplay({ vectors, tComputed }: VectorDisplayProps) {
  if (!vectors) {
    return (
      <p className="vector-display--empty" aria-label="vectors not available">
        Vectors s and e not generated yet.
      </p>
    );
  }

  const { k, s, e } = vectors;

  return (
    <div className="vector-display" aria-label="polynomial vectors">
      <section aria-label="secret vector s">
        {Array.from({ length: k }, (_, i) => (
          <PolyRow key={`s-${i}`} label={`s[${i}]`} poly={s[i] ?? []} />
        ))}
      </section>

      <section aria-label="error vector e">
        {Array.from({ length: k }, (_, i) => (
          <PolyRow key={`e-${i}`} label={`e[${i}]`} poly={e[i] ?? []} />
        ))}
      </section>

      {tComputed && (
        <section aria-label="public vector t">
          {Array.from({ length: tComputed.k }, (_, i) => (
            <PolyRow key={`t-${i}`} label={`t[${i}]`} poly={tComputed.t[i] ?? []} />
          ))}
        </section>
      )}
    </div>
  );
}
