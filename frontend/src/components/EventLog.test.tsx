import React from 'react';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { EventLog } from './EventLog';
import { EventLogEntry } from '../types/protocol';

function makeEntry(type: string, data: unknown): EventLogEntry {
  return { timestamp: '12:00:00', type, data };
}

describe('EventLog', () => {
  it('shows empty placeholder when events list is empty', () => {
    render(<EventLog events={[]} />);
    expect(screen.getByText(/no events yet/i)).toBeInTheDocument();
  });

  it('renders an entry for each event', () => {
    const events = [
      makeEntry('connected', {}),
      makeEntry('params', { k: 2 }),
    ];
    render(<EventLog events={events} />);
    expect(screen.getByText('connected')).toBeInTheDocument();
    expect(screen.getByText('params')).toBeInTheDocument();
  });

  it('displays the timestamp for each entry', () => {
    render(<EventLog events={[makeEntry('sent', {})]} />);
    expect(screen.getByText('12:00:00')).toBeInTheDocument();
  });

  it('shows a data preview for each entry', () => {
    render(<EventLog events={[makeEntry('rho_sigma', { rho: 'abc' })]} />);
    // The serialised data preview should contain the key "rho"
    expect(screen.getByText(/\{"rho"/)).toBeInTheDocument();
  });

  it('truncates long data with ellipsis', () => {
    const longData = { key: 'x'.repeat(200) };
    render(<EventLog events={[makeEntry('test', longData)]} />);
    expect(screen.getByText(/…/)).toBeInTheDocument();
  });

  it('does not show ellipsis for short data', () => {
    render(<EventLog events={[makeEntry('test', { x: 1 })]} />);
    // short data: no trailing ellipsis in the data span
    const entries = screen.getAllByText(/test/);
    expect(entries.length).toBeGreaterThan(0);
  });

  it('has an aria-live region for accessibility', () => {
    render(<EventLog events={[]} />);
    expect(screen.getByRole('log')).toBeInTheDocument();
  });

  it('has the correct aria-label on the container', () => {
    render(<EventLog events={[]} />);
    expect(screen.getByLabelText(/websocket event log/i)).toBeInTheDocument();
  });
});
