import React from 'react';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { VectorDisplay } from './VectorDisplay';
import { VectorsPayload, TComputedPayload } from '../types/protocol';

function makeVec(k: number): VectorsPayload {
  const poly = Array.from({ length: 256 }, (_, i) => i);
  return {
    k,
    s: Array.from({ length: 4 }, () => poly),
    e: Array.from({ length: 4 }, () => poly),
  };
}

function makeT(k: number): TComputedPayload {
  const poly = Array.from({ length: 256 }, (_, i) => i * 2);
  return { k, t: Array.from({ length: 4 }, () => poly) };
}

describe('VectorDisplay', () => {
  it('shows placeholder when vectors is null', () => {
    render(<VectorDisplay vectors={null} tComputed={null} />);
    expect(screen.getByLabelText(/vectors not available/i)).toBeInTheDocument();
  });

  it('renders s and e labels for k=2', () => {
    render(<VectorDisplay vectors={makeVec(2)} tComputed={null} />);
    expect(screen.getByText('s[0]')).toBeInTheDocument();
    expect(screen.getByText('s[1]')).toBeInTheDocument();
    expect(screen.getByText('e[0]')).toBeInTheDocument();
    expect(screen.getByText('e[1]')).toBeInTheDocument();
  });

  it('renders t labels when tComputed is provided', () => {
    render(<VectorDisplay vectors={makeVec(2)} tComputed={makeT(2)} />);
    expect(screen.getByText('t[0]')).toBeInTheDocument();
    expect(screen.getByText('t[1]')).toBeInTheDocument();
  });

  it('does not render t section when tComputed is null', () => {
    render(<VectorDisplay vectors={makeVec(2)} tComputed={null} />);
    expect(screen.queryByText('t[0]')).not.toBeInTheDocument();
  });

  it('shows coefficient preview', () => {
    render(<VectorDisplay vectors={makeVec(2)} tComputed={null} />);
    // The first 8 coefficients of our poly are 0,1,...,7
    expect(screen.getAllByText(/\[0, 1, 2/)[0]).toBeInTheDocument();
  });

  it('renders section for secret vector with aria-label', () => {
    render(<VectorDisplay vectors={makeVec(2)} tComputed={null} />);
    expect(screen.getByLabelText(/secret vector s/i)).toBeInTheDocument();
  });
});
