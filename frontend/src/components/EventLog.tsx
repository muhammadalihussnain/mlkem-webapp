import React, { useEffect, useRef } from 'react';
import { EventLogEntry } from '../types/protocol';

/** Maximum number of characters of serialised data shown per event. */
const DATA_PREVIEW_LEN = 80;

interface EventLogProps {
  events: EventLogEntry[];
}

/**
 * EventLog renders a scrolling list of WebSocket events with timestamps.
 * It auto-scrolls to the newest entry whenever the events list grows.
 */
export function EventLog({ events }: EventLogProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (bottomRef.current && typeof bottomRef.current.scrollIntoView === 'function') {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [events.length]);

  return (
    <div className="event-log" aria-label="WebSocket event log" aria-live="polite">
      <h4 className="event-log__title">Event Log</h4>
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
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
