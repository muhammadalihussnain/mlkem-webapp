import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import App from './App';

describe('Basic tests', () => {
  it('should have working math', () => {
    expect(1 + 1).toBe(2);
  });

  it('should have true boolean', () => {
    expect(true).toBe(true);
  });
});

describe('App Component', () => {
  it('renders without crashing', () => {
    render(<App />);
    // Use getAllByText since there are multiple elements with ML-KEM
    const elements = screen.getAllByText(/ML-KEM/i);
    expect(elements.length).toBeGreaterThan(0);
  });

  it('displays the correct title in heading', () => {
    render(<App />);
    // Look specifically for the heading element
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveTextContent(/ML-KEM/i);
  });

  it('displays ML-KEM in the paragraph', () => {
    render(<App />);
    // Find the paragraph text
    expect(screen.getByText(/CRYSTALS-Kyber/i)).toBeInTheDocument();
  });
});
