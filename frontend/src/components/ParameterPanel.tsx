import React from 'react';
import { ParamsPayload } from '../types/protocol';

interface ParameterPanelProps {
  params: ParamsPayload | null;
}

/**
 * ParameterPanel displays the ML-KEM parameter set in a table.
 * Shows a placeholder message while no flavor has been selected.
 */
export function ParameterPanel({ params }: ParameterPanelProps) {
  if (!params) {
    return (
      <p className="parameter-panel--empty" aria-label="parameter panel empty">
        Select a flavor to see parameters.
      </p>
    );
  }

  const rows: Array<[string, string | number]> = [
    ['n  (ring degree)',       params.n],
    ['q  (modulus)',           params.q],
    ['k  (module rank)',       params.k],
    ['η₁ (key-gen noise)',    params.eta1],
    ['η₂ (encaps noise)',     params.eta2],
    ['dᵤ (u compression)',    params.du],
    ['d𝓿 (v compression)',    params.dv],
    ['|pk| (bytes)',          params.pk_size],
    ['|sk| (bytes)',          params.sk_size],
    ['|ct| (bytes)',          params.ct_size],
  ];

  return (
    <table className="parameter-panel" aria-label="ML-KEM parameters">
      <tbody>
        {rows.map(([label, value]) => (
          <tr key={label}>
            <td>{label}</td>
            <td>{value}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
