import React from 'react';
import { MatrixAPayload } from '../types/protocol';

/** Number of coefficients shown before the ellipsis. */
const PREVIEW_COEFFS = 8;

interface MatrixDisplayProps {
  matrix: MatrixAPayload | null;
}

/**
 * MatrixDisplay renders the public NTT-domain matrix A as a k×k grid.
 * Each cell shows the first PREVIEW_COEFFS coefficients of the polynomial
 * followed by "…" to keep the UI compact.
 */
export function MatrixDisplay({ matrix }: MatrixDisplayProps) {
  if (!matrix) {
    return (
      <p className="matrix-display--empty" aria-label="matrix not available">
        Matrix A not generated yet.
      </p>
    );
  }

  const { k, a } = matrix;

  return (
    <div className="matrix-display" aria-label={`Matrix A ${k} by ${k}`}>
      <p className="matrix-display__dim">A ∈ ℤ_q[X]/(Xⁿ+1)^{'{'}k×k{'}'} — NTT domain</p>
      <div
        className="matrix-display__grid"
        style={{ gridTemplateColumns: `repeat(${k}, 1fr)` }}
      >
        {Array.from({ length: k }, (_, i) =>
          Array.from({ length: k }, (_, j) => {
            const poly = a[i]?.[j] ?? [];
            const preview = poly.slice(0, PREVIEW_COEFFS).join(', ');
            return (
              <div key={`${i}-${j}`} className="matrix-display__cell">
                <span className="matrix-display__label">A[{i}][{j}]</span>
                <span className="matrix-display__coeffs" title={`Full polynomial A[${i}][${j}]`}>
                  [{preview}…]
                </span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
