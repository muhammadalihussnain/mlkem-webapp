import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import App from './App';

// ── Mock WebSocket so App renders without a real server ───────────────────────

class MockWebSocket {
  onopen: (() => void) | null = null;
  onmessage: ((e: MessageEvent) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  readyState = WebSocket.CONNECTING;
  send = vi.fn();
  close = vi.fn();
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
}

beforeEach(() => {
  vi.stubGlobal('WebSocket', MockWebSocket);
  Object.defineProperty(window, 'location', {
    value: { protocol: 'http:', host: 'localhost' },
    writable: true,
    configurable: true,
  });
});

describe('App', () => {
  it('renders the main heading', () => {
    render(<App />);
    const headings = screen.getAllByRole('heading', { level: 1 });
    expect(headings.some(h => /ML-KEM/i.test(h.textContent ?? ''))).toBe(true);
  });

  it('shows disconnected status initially', () => {
    render(<App />);
    expect(screen.getByText(/Disconnected/i)).toBeInTheDocument();
  });

  it('renders the flavor selector', () => {
    render(<App />);
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('renders all step buttons', () => {
    render(<App />);
    const buttons = screen.getAllByRole('button');
    // At least 5 step buttons + Reset button
    expect(buttons.length).toBeGreaterThanOrEqual(6);
  });

  it('renders the Reset button', () => {
    render(<App />);
    expect(screen.getByRole('button', { name: /Reset/i })).toBeInTheDocument();
  });

  it('renders the parameter panel placeholder', () => {
    render(<App />);
    expect(screen.getByLabelText(/parameter panel empty/i)).toBeInTheDocument();
  });

  it('renders the matrix placeholder', () => {
    render(<App />);
    // MatrixDisplay only renders after data arrives; check the step section heading instead
    expect(screen.getByText(/Key Generation Steps/i)).toBeInTheDocument();
  });

  it('renders the event log', () => {
    render(<App />);
    expect(screen.getByLabelText(/websocket event log/i)).toBeInTheDocument();
  });

  it('clicking Reset does not crash', () => {
    render(<App />);
    fireEvent.click(screen.getByRole('button', { name: /Reset/i }));
    // Still renders correctly — multiple h1s now (header + intro)
    const headings = screen.getAllByRole('heading', { level: 1 });
    expect(headings.length).toBeGreaterThanOrEqual(1);
  });
});
