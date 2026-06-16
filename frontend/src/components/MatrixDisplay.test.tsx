import React from 'react';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MatrixDisplay } from './MatrixDisplay';
import { MatrixAPayload } from '../types/protocol';

function makeMatrix(k: number): MatrixAPayload {
  const a: number[][][] = Array.from({ length: 4 }, (_, i) =>
    Array.from({ length: 4 }, (_, j) =>
      Array.from({ length: 256 }, (_, c) => (i * 4 + j) * 10 + c)
    )
  );
  return { k, a };
}

describe('MatrixDisplay', () => {
  it('shows placeholder when matrix is null', () => {
    render(<MatrixDisplay matrix={null} />);
    expect(screen.getByLabelText(/matrix not available/i)).toBeInTheDocument();
  });

  it('renders k×k cells for k=2', () => {
    render(<MatrixDisplay matrix={makeMatrix(2)} />);
    // 4 labels: A[0][0], A[0][1], A[1][0], A[1][1]
    expect(screen.getByText('A[0][0]')).toBeInTheDocument();
    expect(screen.getByText('A[1][1]')).toBeInTheDocument();
  });

  it('renders k×k cells for k=3', () => {
    render(<MatrixDisplay matrix={makeMatrix(3)} />);
    expect(screen.getByText('A[2][2]')).toBeInTheDocument();
  });

  it('renders k×k cells for k=4', () => {
    render(<MatrixDisplay matrix={makeMatrix(4)} />);
    expect(screen.getByText('A[3][3]')).toBeInTheDocument();
  });

  it('shows coefficient preview with ellipsis', () => {
    render(<MatrixDisplay matrix={makeMatrix(2)} />);
    // Coefficients start at 0 for A[0][0]
    const cells = screen.getAllByText(/\[.*…\]/);
    expect(cells.length).toBeGreaterThan(0);
  });

  it('has correct aria-label', () => {
    render(<MatrixDisplay matrix={makeMatrix(3)} />);
    expect(screen.getByLabelText(/Matrix A 3 by 3/i)).toBeInTheDocument();
  });
});
