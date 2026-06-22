import { useState } from 'react';
import type { MatrixAPayload } from '../types/protocol';

const PREVIEW_COEFFS = 8;

interface MatrixDisplayProps {
  matrix: MatrixAPayload | null;
}

interface CellTooltipProps {
  label: string;
  poly: number[];
  onClose: () => void;
}

function CellTooltip({ label, poly, onClose }: CellTooltipProps) {
  const all = poly.join(', ');
  const preview = poly.slice(0, 32).join(', ');
  const truncated = poly.length > 32;
  return (
    <div className="matrix-tooltip" role="dialog" aria-label={`Polynomial ${label}`}>
      <div className="matrix-tooltip__header">
        <span className="matrix-tooltip__label">{label}</span>
        <button className="matrix-tooltip__close" onClick={onClose} aria-label="close">✕</button>
      </div>
      <div className="matrix-tooltip__meta">
        {poly.length} coefficients · NTT domain · mod q = 3329
      </div>
      <div className="matrix-tooltip__desc">
        Each value is a polynomial coefficient in Z_q = &#123;0,…,3328&#125; in the NTT (frequency) domain.
        These are produced by SHAKE-128 rejection sampling from the public seed ρ.
      </div>
      <div className="matrix-tooltip__coeffs" title={all}>
        [{preview}{truncated ? '…' : ''}]
      </div>
    </div>
  );
}

export function MatrixDisplay({ matrix }: MatrixDisplayProps) {
  const [activeCell, setActiveCell] = useState<string | null>(null);

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
      <p className="matrix-display__dim">
        A &#8712; &#8484;_q[X]/(X&#8319;+1)&#x7B;k&times;k&#x7D; &mdash; NTT domain
      </p>
      <div
        className="matrix-display__grid"
        style={{ gridTemplateColumns: `repeat(${k}, 1fr)` }}
      >
        {Array.from({ length: k }, (_, i) =>
          Array.from({ length: k }, (_, j) => {
            const cellId = `${i}-${j}`;
            const poly = a[i]?.[j] ?? [];
            const preview = poly.slice(0, PREVIEW_COEFFS).join(', ');
            const isActive = activeCell === cellId;
            return (
              <div key={cellId} className={`matrix-display__cell${isActive ? ' matrix-display__cell--active' : ''}`}
                style={{ position: 'relative' }}>
                <button
                  className="matrix-display__cell-btn"
                  onClick={() => setActiveCell(isActive ? null : cellId)}
                  aria-expanded={isActive}
                  aria-label={`Show details for A[${i}][${j}]`}
                >
                  <span className="matrix-display__label">A[{i}][{j}]</span>
                  <span className="matrix-display__coeffs">
                    [{preview}…]
                  </span>
                </button>
                {isActive && (
                  <CellTooltip
                    label={`A[${i}][${j}]`}
                    poly={poly}
                    onClose={() => setActiveCell(null)}
                  />
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
