// ── Inbound message types (frontend → backend) ────────────────────────────────

export interface SelectFlavorMessage {
  type: 'select_flavor';
  flavor: string; // "512" | "768" | "1024"
}

export interface StepNextMessage {
  type: 'step_next';
  step: string;
}

export interface SendMessageMsg {
  type: 'send_message';
}

export type OutgoingMessage = SelectFlavorMessage | StepNextMessage | SendMessageMsg;

// ── Backend payload shapes (inside OutboundMessage.payload) ───────────────────

export interface ParamsPayload {
  flavor: string;
  n: number;
  q: number;
  k: number;
  eta1: number;
  eta2: number;
  du: number;
  dv: number;
  pk_size: number;
  sk_size: number;
  ct_size: number;
}

export interface RhoSigmaPayload {
  seed: string;  // hex
  rho: string;   // hex
  sigma: string; // hex
}

export interface MatrixAPayload {
  k: number;
  a: number[][][]; // [4][4][256] but only k×k populated
}

export interface VectorsPayload {
  k: number;
  s: number[][];  // [4][256]
  e: number[][];  // [4][256]
}

export interface TComputedPayload {
  k: number;
  t: number[][];  // [4][256]
}

export interface PublicKeyPayload {
  public_key: string;      // hex
  public_key_size: number; // bytes
}

export interface PrivateKeyPayload {
  private_key: string;      // hex
  private_key_size: number; // bytes
}

export interface EncryptResultPayload {
  ciphertext: string;      // hex
  ciphertext_size: number; // bytes
  shared_secret: string;   // hex
  message: string;         // hex (random plaintext embedded in ct)
}

export interface DecryptResultPayload {
  shared_secret: string; // hex
  match: boolean;
}

export interface ErrorPayload {
  message: string;
}

// ── WebSocket message envelope ────────────────────────────────────────────────

export interface WsMessage {
  type: string;
  payload: unknown;
}

// ── App state ─────────────────────────────────────────────────────────────────

export interface AppState {
  flavor: string;
  params: ParamsPayload | null;
  rhoSigma: RhoSigmaPayload | null;
  matrixA: MatrixAPayload | null;
  vectors: VectorsPayload | null;
  tComputed: TComputedPayload | null;
  publicKey: PublicKeyPayload | null;
  privateKey: PrivateKeyPayload | null;
  encryptResult: EncryptResultPayload | null;
  decryptResult: DecryptResultPayload | null;
  currentStep: number;
  events: EventLogEntry[];
}

export interface EventLogEntry {
  timestamp: string;
  type: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: any;
}

// ── Step definitions ──────────────────────────────────────────────────────────

export interface Step {
  id: string;
  label: string;
}

export const STEPS: Step[] = [
  { id: 'generate_rho_sigma', label: 'Generate ρ / σ' },
  { id: 'generate_matrix_A', label: 'Generate Matrix A' },
  { id: 'generate_vectors',  label: 'Generate s / e' },
  { id: 'compute_t',         label: 'Compute t = A·s + e' },
  { id: 'send_public_key',   label: 'Encode Public Key' },
];

// ── Flavor options ────────────────────────────────────────────────────────────

export const FLAVORS = [
  { value: '512',  label: 'ML-KEM-512' },
  { value: '768',  label: 'ML-KEM-768' },
  { value: '1024', label: 'ML-KEM-1024' },
] as const;

export type FlavorValue = typeof FLAVORS[number]['value'];
