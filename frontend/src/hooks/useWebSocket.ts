import { useCallback, useEffect, useRef, useState } from 'react';
import {
  AppState,
  DecryptResultPayload,
  EncryptResultPayload,
  MatrixAPayload,
  OutgoingMessage,
  ParamsPayload,
  PrivateKeyPayload,
  PublicKeyPayload,
  RhoSigmaPayload,
  TComputedPayload,
  VectorsPayload,
  WsMessage,
} from '../types/protocol';

/** WebSocket endpoint — proxied by Vite in dev, direct in production. */
const WS_URL = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`;

/** How long to wait before attempting a reconnect (ms). */
const RECONNECT_DELAY_MS = 3000;

const INITIAL_STATE: AppState = {
  flavor: '',
  params: null,
  rhoSigma: null,
  matrixA: null,
  vectors: null,
  tComputed: null,
  publicKey: null,
  privateKey: null,
  encryptResult: null,
  decryptResult: null,
  currentStep: 0,
  events: [],
};

export function useWebSocket() {
  const [isConnected, setIsConnected] = useState(false);
  const [state, setState] = useState<AppState>(INITIAL_STATE);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // ── Event log helper ───────────────────────────────────────────────────────

  const addEvent = useCallback((type: string, data: unknown) => {
    setState(prev => ({
      ...prev,
      events: [
        ...prev.events,
        { timestamp: new Date().toLocaleTimeString(), type, data },
      ],
    }));
  }, []);

  // ── Message handler ────────────────────────────────────────────────────────

  const handleMessage = useCallback((msg: WsMessage) => {
    setState(prev => {
      const next = { ...prev };

      switch (msg.type) {
        case 'params':
          next.params = msg.payload as ParamsPayload;
          next.flavor = (msg.payload as ParamsPayload).flavor;
          next.currentStep = 1;
          break;

        case 'rho_sigma':
          next.rhoSigma = msg.payload as RhoSigmaPayload;
          next.currentStep = 2;
          break;

        case 'matrix_A':
          next.matrixA = msg.payload as MatrixAPayload;
          next.currentStep = 3;
          break;

        case 'vectors':
          next.vectors = msg.payload as VectorsPayload;
          next.currentStep = 4;
          break;

        case 't_computed':
          next.tComputed = msg.payload as TComputedPayload;
          next.currentStep = 5;
          break;

        case 'public_key_sent':
          next.publicKey = msg.payload as PublicKeyPayload;
          next.currentStep = 6;
          break;

        case 'public_key_recv':
          next.privateKey = msg.payload as PrivateKeyPayload;
          break;

        case 'encrypt_result':
          next.encryptResult = msg.payload as EncryptResultPayload;
          next.currentStep = 7;
          break;

        case 'decrypt_result':
          next.decryptResult = msg.payload as DecryptResultPayload;
          next.currentStep = 8;
          break;

        // 'error' and unknown types are logged but don't mutate state.
      }

      return next;
    });
  }, []);

  // ── Send ───────────────────────────────────────────────────────────────────

  const sendMessage = useCallback((message: OutgoingMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
      addEvent('sent', message);
    } else {
      console.warn('ws: not connected, message dropped', message);
    }
  }, [addEvent]);

  // ── Connect / disconnect ───────────────────────────────────────────────────

  const connect = useCallback(() => {
    if (wsRef.current) return; // already connecting or connected

    try {
      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        addEvent('connected', {});
      };

      ws.onmessage = (event: MessageEvent) => {
        try {
          const msg = JSON.parse(event.data as string) as WsMessage;
          handleMessage(msg);
          addEvent(msg.type, msg.payload);
        } catch (err) {
          console.error('ws: parse error', err);
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        wsRef.current = null;
        addEvent('disconnected', {});
        reconnectRef.current = setTimeout(connect, RECONNECT_DELAY_MS);
      };

      ws.onerror = () => {
        // onclose fires after onerror; no extra state update needed.
        addEvent('error', {});
      };
    } catch (err) {
      console.error('ws: connect failed', err);
    }
  }, [addEvent, handleMessage]);

  const disconnect = useCallback(() => {
    if (reconnectRef.current) {
      clearTimeout(reconnectRef.current);
      reconnectRef.current = null;
    }
    wsRef.current?.close();
    wsRef.current = null;
    setIsConnected(false);
  }, []);

  const resetState = useCallback(() => {
    setState(INITIAL_STATE);
  }, []);

  // ── Lifecycle ──────────────────────────────────────────────────────────────

  useEffect(() => {
    connect();
    return disconnect;
  }, [connect, disconnect]);

  return { isConnected, state, sendMessage, disconnect, resetState };
}
