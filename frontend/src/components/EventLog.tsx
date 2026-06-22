import React from 'react';
import type { EventLogEntry } from '../types/protocol';

/** Maximum number of characters of serialised data shown per event. */
const DATA_PREVIEW_LEN = 80;

interface EventLogProps {
  events: EventLogEntry[];
}

export function EventLog({ events }: EventLogProps) {
  return (
    <div className="event-log" aria-label="WebSocket event log" aria-live="off">
      <div className="event-log__entries" role="log">
        {events.length === 0 ? (
          <p className="event-log__empty">No events yet.</p>
        ) : (
          events.map((entry, idx) => {
            const preview = JSON.stringify(entry.data).slice(0, DATA_PREVIEW_LEN);
            const truncated = JSON.stringify(entry.data).length > DATA_PREVIEW_LEN;
            return (
              <div key={idx} className={`event-log__entry event-log__entry--${entry.type}`}>
                <span className="event-log__time">{entry.timestamp}</span>
                <span className="event-log__type">{entry.type}</span>
                <span className="event-log__data">{preview}{truncated ? '…' : ''}</span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
