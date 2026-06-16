import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useWebSocket } from './useWebSocket';

// ── Mock WebSocket ─────────────────────────────────────────────────────────────

type WsHandler = (event: Event | MessageEvent | CloseEvent) => void;

interface MockWsInstance {
  onopen: WsHandler | null;
  onmessage: WsHandler | null;
  onclose: WsHandler | null;
  onerror: WsHandler | null;
  readyState: number;
  send: ReturnType<typeof vi.fn>;
  close: ReturnType<typeof vi.fn>;
  // helpers for tests
  simulateOpen: () => void;
  simulateMessage: (data: unknown) => void;
  simulateClose: () => void;
  simulateError: () => void;
}

let mockInstance: MockWsInstance;

class MockWebSocket {
  onopen: WsHandler | null = null;
  onmessage: WsHandler | null = null;
  onclose: WsHandler | null = null;
  onerror: WsHandler | null = null;
  readyState = WebSocket.CONNECTING;
  send = vi.fn();
  close = vi.fn().mockImplementation(() => {
    this.readyState = WebSocket.CLOSED;
  });

  constructor(_url: string) {
    mockInstance = this as unknown as MockWsInstance;
    (mockInstance as unknown as MockWebSocket).simulateOpen = () => {
      this.readyState = WebSocket.OPEN;
      this.onopen?.(new Event('open'));
    };
    (mockInstance as unknown as MockWebSocket).simulateMessage = (data: unknown) => {
      this.onmessage?.(new MessageEvent('message', { data: JSON.stringify(data) }));
    };
    (mockInstance as unknown as MockWebSocket).simulateClose = () => {
      this.readyState = WebSocket.CLOSED;
      this.onclose?.(new CloseEvent('close'));
    };
    (mockInstance as unknown as MockWebSocket).simulateError = () => {
      this.onerror?.(new Event('error'));
    };
  }

  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
}

// ── Setup / teardown ───────────────────────────────────────────────────────────

beforeEach(() => {
  vi.stubGlobal('WebSocket', MockWebSocket);
  // Stub window.location so WS_URL construction doesn't throw
  Object.defineProperty(window, 'location', {
    value: { protocol: 'http:', host: 'localhost' },
    writable: true,
    configurable: true,
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.clearAllTimers();
});

// ── Tests ──────────────────────────────────────────────────────────────────────

describe('useWebSocket', () => {
  it('starts disconnected', () => {
    const { result } = renderHook(() => useWebSocket());
    expect(result.current.isConnected).toBe(false);
  });

  it('becomes connected after WebSocket opens', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    expect(result.current.isConnected).toBe(true);
  });

  it('adds a connected event on open', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    expect(result.current.state.events.some(e => e.type === 'connected')).toBe(true);
  });

  it('becomes disconnected after WebSocket closes', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => { mockInstance.simulateClose(); });
    expect(result.current.isConnected).toBe(false);
  });

  it('handles params message and updates state', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({
        type: 'params',
        payload: { flavor: '512', n: 256, q: 3329, k: 2, eta1: 3, eta2: 2, du: 10, dv: 4, pk_size: 800, sk_size: 768, ct_size: 768 },
      });
    });
    expect(result.current.state.params).not.toBeNull();
    expect(result.current.state.params?.k).toBe(2);
    expect(result.current.state.currentStep).toBe(0);
  });

  it('handles rho_sigma message', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({
        type: 'rho_sigma',
        payload: { seed: 'aa', rho: 'bb', sigma: 'cc' },
      });
    });
    expect(result.current.state.rhoSigma?.rho).toBe('bb');
    expect(result.current.state.currentStep).toBe(1);
  });

  it('handles matrix_A message', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({ type: 'matrix_A', payload: { k: 2, a: [] } });
    });
    expect(result.current.state.matrixA?.k).toBe(2);
    expect(result.current.state.currentStep).toBe(2);
  });

  it('handles vectors message', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({ type: 'vectors', payload: { k: 2, s: [], e: [] } });
    });
    expect(result.current.state.vectors).not.toBeNull();
    expect(result.current.state.currentStep).toBe(3);
  });

  it('handles t_computed message', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({ type: 't_computed', payload: { k: 2, t: [] } });
    });
    expect(result.current.state.tComputed).not.toBeNull();
    expect(result.current.state.currentStep).toBe(4);
  });

  it('handles public_key_sent message', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({ type: 'public_key_sent', payload: { public_key: 'abc', public_key_size: 800 } });
    });
    expect(result.current.state.publicKey?.public_key_size).toBe(800);
    expect(result.current.state.currentStep).toBe(5);
  });

  it('handles public_key_recv message', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({ type: 'public_key_recv', payload: { private_key: 'sk', private_key_size: 768 } });
    });
    expect(result.current.state.privateKey?.private_key_size).toBe(768);
  });

  it('handles encrypt_result message', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({
        type: 'encrypt_result',
        payload: { ciphertext: 'ct', ciphertext_size: 768, shared_secret: 'ss', message: 'mm' },
      });
    });
    expect(result.current.state.encryptResult?.ciphertext_size).toBe(768);
    expect(result.current.state.currentStep).toBe(6);
  });

  it('handles decrypt_result message', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({ type: 'decrypt_result', payload: { shared_secret: 'ss', match: true } });
    });
    expect(result.current.state.decryptResult?.match).toBe(true);
    expect(result.current.state.currentStep).toBe(7);
  });

  it('ignores unknown message types without throwing', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({ type: 'unknown_type', payload: {} });
    });
    expect(result.current.state.currentStep).toBe(0);
  });

  it('sendMessage sends JSON when connected', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      result.current.sendMessage({ type: 'select_flavor', flavor: '512' });
    });
    expect(mockInstance.send).toHaveBeenCalledWith(
      JSON.stringify({ type: 'select_flavor', flavor: '512' })
    );
  });

  it('sendMessage warns when not connected', () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
    const { result } = renderHook(() => useWebSocket());
    // Don't open — readyState stays CONNECTING
    act(() => {
      result.current.sendMessage({ type: 'select_flavor', flavor: '512' });
    });
    expect(warn).toHaveBeenCalled();
    warn.mockRestore();
  });

  it('resetState clears all state', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => {
      mockInstance.simulateMessage({ type: 'matrix_A', payload: { k: 3, a: [] } });
    });
    act(() => { result.current.resetState(); });
    expect(result.current.state.matrixA).toBeNull();
    expect(result.current.state.currentStep).toBe(0);
  });

  it('disconnect closes the WebSocket', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => { result.current.disconnect(); });
    expect(mockInstance.close).toHaveBeenCalled();
  });

  it('adds error event on WebSocket error', () => {
    const { result } = renderHook(() => useWebSocket());
    act(() => { mockInstance.simulateOpen(); });
    act(() => { mockInstance.simulateError(); });
    expect(result.current.state.events.some(e => e.type === 'error')).toBe(true);
  });
});
