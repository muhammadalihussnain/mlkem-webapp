import React from 'react';
import { FLAVORS } from '../types/protocol';

interface FlavorSelectorProps {
  /** Called when the user picks a flavor. Receives the backend value ("512" etc.). */
  onSelect: (flavor: string) => void;
  /** Currently selected flavor value, or empty string for none. */
  currentFlavor: string;
  /** When true the control is read-only. */
  disabled: boolean;
}

/**
 * FlavorSelector renders a labelled `<select>` that lets the user choose an
 * ML-KEM security level and immediately fires `onSelect` with the chosen value.
 */
export function FlavorSelector({ onSelect, currentFlavor, disabled }: FlavorSelectorProps) {
  return (
    <div className="flavor-selector">
      <label htmlFor="flavor-select">Security level</label>
      <select
        id="flavor-select"
        value={currentFlavor}
        onChange={e => onSelect(e.target.value)}
        disabled={disabled}
        aria-label="ML-KEM security level"
      >
        <option value="">— select —</option>
        {FLAVORS.map(f => (
          <option key={f.value} value={f.value}>
            {f.label}
          </option>
        ))}
      </select>
    </div>
  );
}
