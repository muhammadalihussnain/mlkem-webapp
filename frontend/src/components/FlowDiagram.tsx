import { useState } from 'react';

// ── Types ────────────────────────────────────────────────────────────────────
type NodeColor = 'neutral' | 'green' | 'red';

interface Node {
  id: string; label: string; sub: string; color: NodeColor;
  cx: number; cy: number; w: number; h: number;
  tooltip: string;
  definition: string;
  wikiUrl: string;
  specUrl: string;
  specLabel: string;
  step: number;
}
interface Arrow { from: string; to: string; label?: string; dashed?: boolean; }

// ── Colour palette — pure black text on clean backgrounds ───────────────────
const NODE_STROKE: Record<NodeColor, string> = {
  neutral: '#1e293b',   // dark slate border
  green:   '#065f46',   // emerald
  red:     '#9f1239',   // rose
};
const NODE_FILL: Record<NodeColor, string> = {
  neutral: '#ffffff',
  green:   '#ecfdf5',
  red:     '#fff1f2',
};
const NODE_LABEL: Record<NodeColor, string> = {
  neutral: '#000000',   // pure black
  green:   '#064e3b',
  red:     '#881337',
};
const NODE_SUB: Record<NodeColor, string> = {
  neutral: '#374151',   // dark grey — clearly readable
  green:   '#065f46',
  red:     '#9f1239',
};

// ── Tooltip colours (theme-aware) ────────────────────────────────────────────
function cssVar(name: string, fallback: string) {
  if (typeof document === 'undefined') return fallback;
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || fallback;
}

// ── Node data ────────────────────────────────────────────────────────────────
// viewBox: 960 × 1020  — nodes are larger and better spaced
const NODES: Node[] = [
  {
    id: 'csprng', label: 'CSPRNG', sub: 'Cryptographically Secure Random Generator',
    color: 'neutral', cx: 480, cy: 60, w: 280, h: 56, step: 1,
    tooltip: 'Takes entropy from the OS (/dev/urandom) or hardware RNG. Produces two independent 32-byte values: d (seed for hashing) and z (implicit-rejection secret). This is the only true source of randomness — all subsequent steps are deterministic.',
    definition: 'output ∈ {0,1}^n, P[next bit=1 | all prior bits] = ½. FIPS 203 requires a NIST-approved DRBG (SP 800-90A) or OS entropy source.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Cryptographically_secure_pseudorandom_number_generator',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §3.3',
  },
  {
    id: 'd', label: 'd  (32 bytes)', sub: 'Primary seed → fed into G = SHA3-512',
    color: 'neutral', cx: 300, cy: 170, w: 220, h: 52, step: 1,
    tooltip: 'First 32-byte output of CSPRNG. Fed directly into G = SHA3-512(d) as the sole input. Must stay secret — its exposure is equivalent to losing the entire key pair.',
    definition: 'd ∈ {0,1}^256 — primary 256-bit entropy seed. All key material is deterministically derived from d.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Random_seed',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §5.1',
  },
  {
    id: 'z', label: 'z  (32 bytes)', sub: 'Implicit-rejection secret → stored in dk',
    color: 'neutral', cx: 720, cy: 170, w: 240, h: 52, step: 1,
    tooltip: 'Second 32-byte output of CSPRNG, independent of d. Stored in the decapsulation key. If ciphertext tampering is detected, the output becomes H(z‖c) instead of the real secret — preventing timing side-channel attacks.',
    definition: 'z ∈ {0,1}^256 — implicit rejection value. Ensures IND-CCA2 security without revealing whether decapsulation succeeded.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Key_encapsulation_mechanism',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §5.1 Alg.19',
  },
  {
    id: 'G', label: 'G = SHA3-512(d)', sub: 'Hash expands 32 bytes → 64 bytes (ρ ‖ σ)',
    color: 'neutral', cx: 300, cy: 284, w: 280, h: 56, step: 2,
    tooltip: 'SHA3-512 hashes the seed d and produces 64 bytes. The first 32 bytes become ρ (public matrix seed) and the last 32 bytes become σ (private noise seed). Being a one-way function, you cannot recover d from ρ and σ.',
    definition: 'G: {0,1}^256 → {0,1}^512.  G(d) = (ρ ‖ σ),  ρ = G(d)[0..31],  σ = G(d)[32..63].  FIPS 203 §4.1.',
    wikiUrl: 'https://en.wikipedia.org/wiki/SHA-3',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §4.1',
  },
  {
    id: 'rho', label: 'ρ  (32 bytes)', sub: 'Public matrix seed — included in ek',
    color: 'neutral', cx: 180, cy: 392, w: 210, h: 52, step: 2,
    tooltip: 'Bytes 0–31 of G(d). Public — included in the encapsulation key so any party can deterministically reconstruct the matrix A. Fed into SHAKE-128 to expand into A.',
    definition: 'ρ ∈ {0,1}^256 — public seed.  Â[i][j] = SampleNTT(ρ, i, j)  for i,j ∈ [k].',
    wikiUrl: 'https://en.wikipedia.org/wiki/Module_Learning_with_Errors',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §5.1',
  },
  {
    id: 'sigma', label: 'σ  (32 bytes)', sub: 'Private noise seed — never leaves dk',
    color: 'neutral', cx: 460, cy: 392, w: 220, h: 52, step: 2,
    tooltip: 'Bytes 32–63 of G(d). Private — must never be revealed. Fed into SHAKE-256 (PRF) to generate the secret vector s and error vector e via the CBD sampler.',
    definition: 'σ ∈ {0,1}^256 — private PRF seed.  s = CBD(PRF(σ,0..k-1)),  e = CBD(PRF(σ,k..2k-1)).',
    wikiUrl: 'https://en.wikipedia.org/wiki/Pseudorandom_function_family',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §5.1',
  },
  {
    id: 'shake128', label: 'SHAKE-128  (XOF)', sub: 'Input: ρ ‖ row i ‖ col j',
    color: 'neutral', cx: 180, cy: 496, w: 210, h: 52, step: 3,
    tooltip: 'An eXtendable-Output Function fed with ρ plus the row and column indices. Produces an unbounded pseudorandom byte stream that is rejection-sampled into uniform polynomial coefficients in Z_q. Called k² times — once per matrix entry.',
    definition: 'SHAKE-128: {0,1}^* → {0,1}^∞.  Called k² times to fill each Â[i][j].  FIPS 202 §6.2.',
    wikiUrl: 'https://en.wikipedia.org/wiki/SHAKE128',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.202', specLabel: 'FIPS 202 §6.2',
  },
  {
    id: 'prf', label: 'PRF = SHAKE-256', sub: 'Input: σ ‖ counter N',
    color: 'neutral', cx: 460, cy: 496, w: 210, h: 52, step: 4,
    tooltip: 'SHAKE-256 used as a Pseudorandom Function. Takes σ and an incrementing counter N. Produces byte streams that are fed into the CBD sampler — k times to build s and k times to build e.',
    definition: 'PRF(σ,N) = SHAKE-256(σ ‖ N)[0..64η−1].  N ∈ {0,1}^8 distinguishes each polynomial.  FIPS 203 §4.1.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Pseudorandom_function_family',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §4.1',
  },
  {
    id: 'sampleNTT', label: 'SampleNTT', sub: 'Rejection-sample XOF stream → uniform Z_q poly',
    color: 'neutral', cx: 180, cy: 604, w: 210, h: 52, step: 3,
    tooltip: 'Reads bytes from the SHAKE-128 stream and keeps only values less than q = 3329. Each accepted value becomes one coefficient of a polynomial. The result is already in NTT domain — no separate NTT transform is needed for A.',
    definition: 'SampleNTT: {0,1}^* → R_q.  Rejects bytes ≥ q=3329.  Returns â ∈ Z_q^{256} in NTT domain.  FIPS 203 Alg.7.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Rejection_sampling',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §4.2.1 Alg.7',
  },
  {
    id: 'cbd', label: 'CBD  (η-sampler)', sub: 'Centered Binomial Distribution — small coefficients',
    color: 'neutral', cx: 460, cy: 604, w: 220, h: 52, step: 4,
    tooltip: 'Converts PRF bytes into polynomials with small coefficients in {−η,…,+η}. Each coefficient is computed as (sum of η random bits) minus (sum of η more random bits). For ML-KEM-512, η₁ = 3, giving coefficients in {−3,…,3}.',
    definition: 'CBD_η: {0,1}^{64η} → R_q.  coeff_i = Σ_{j<η} a_{2ηi+j} − Σ_{j<η} a_{2ηi+η+j}.  ‖coeff‖_∞ ≤ η.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Binomial_distribution',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §4.2.2 Alg.8',
  },
  {
    id: 'matA', label: 'Matrix  Â  (k × k)', sub: 'Public matrix — already in NTT domain',
    color: 'neutral', cx: 180, cy: 716, w: 230, h: 56, step: 3,
    tooltip: 'The k×k matrix of polynomials produced directly in NTT domain by SampleNTT. For ML-KEM-512 (k=2) this is a 2×2 matrix — 4 polynomials each of degree 255 with coefficients mod q=3329. This matrix is public and included in the encapsulation key.',
    definition: 'Â ∈ R_q^{k×k}.  Â[i][j] = SampleNTT(ρ,i,j).  256 coefficients ∈ Z_{3329} per entry.  Public.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Module_Learning_with_Errors',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §5.1 Alg.13',
  },
  {
    id: 's', label: 's  (secret vector)', sub: 'k polynomials with small coefficients',
    color: 'neutral', cx: 480, cy: 716, w: 220, h: 56, step: 4,
    tooltip: 'A vector of k polynomials with small coefficients sampled from the CBD distribution. This is the core private key. Before it can be multiplied with A, it must be converted to the NTT domain.',
    definition: 's ∈ R_q^k.  s[i] = CBD_{η₁}(PRF(σ,i)).  ‖s‖_∞ ≤ η₁.  Core secret.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Learning_with_errors',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §5.1 Alg.13',
  },
  {
    id: 'e', label: 'e  (error vector)', sub: 'Small-coefficient noise — stays in standard domain',
    color: 'neutral', cx: 760, cy: 716, w: 230, h: 56, step: 4,
    tooltip: 'A vector of k polynomials with small coefficients. Unlike s, e is NEVER transformed — it stays in the standard polynomial domain and is added directly after the inverse NTT. This noise term is what makes the Module-LWE problem hard.',
    definition: 'e ∈ R_q^k.  e[i] = CBD_{η₁}(PRF(σ,k+i)).  Added in standard domain after INTT.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Learning_with_errors',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §5.1 Alg.13',
  },
  {
    id: 'ntts', label: 'NTT(s)  =  ŝ', sub: 'Number Theoretic Transform — O(n log n)',
    color: 'neutral', cx: 480, cy: 824, w: 220, h: 56, step: 5,
    tooltip: 'Applies the Number Theoretic Transform to s, converting it to the NTT domain (ŝ). This allows multiplication with Â to be done coefficient-wise (Hadamard product) instead of full polynomial convolution, reducing complexity from O(n²) to O(n log n).',
    definition: 'ŝ = NTT(s) ∈ R_q^k.  NTT(f)[i] = Σ_{j=0}^{255} f_j · ω^{ij} mod q,  ω = primitive 512th root of unity mod 3329.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Number_theoretic_transform',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §4.3 Alg.9',
  },
  {
    id: 't', label: 't  =  INTT( Â ⊗ ŝ )  +  e', sub: 'Public LWE vector — combines all key material',
    color: 'neutral', cx: 480, cy: 932, w: 360, h: 60, step: 6,
    tooltip: 'The public vector t is computed by: (1) multiplying Â and ŝ pointwise in the NTT domain, (2) applying the inverse NTT (INTT) to return to standard domain, then (3) adding the small error vector e. Given Â and t, recovering s is computationally infeasible due to e — this is the Module-LWE hardness assumption.',
    definition: 't = INTT(Â ⊗ ŝ) + e ∈ R_q^k.  Module-LWE: t ≈ A·s (mod q).  Security ↔ MLWE_{k,η₁} hardness.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Module_Learning_with_Errors',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §5.1 Alg.13',
  },
  {
    id: 'ek', label: 'Encapsulation Key  ek', sub: 'ByteEncode(t) ‖ ρ  —  PUBLIC KEY',
    color: 'green', cx: 300, cy: 1040, w: 270, h: 60, step: 7,
    tooltip: 'The public encapsulation key. Formed by serialising (byte-encoding) the public vector t and appending ρ. Anyone can use this key to encapsulate a shared secret. Sizes: 800 bytes (ML-KEM-512), 1184 bytes (ML-KEM-768), 1568 bytes (ML-KEM-1024).',
    definition: 'ek = ByteEncode_{dt}(t) ‖ ρ.  |ek| = 800/1184/1568 bytes for k=2/3/4.  FIPS 203 §5.1 Alg.13.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Key_encapsulation_mechanism',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §5.1 Alg.13',
  },
  {
    id: 'dk', label: 'Decapsulation Key  dk', sub: 'ByteEncode(s) ‖ ek ‖ H(ek) ‖ z  —  PRIVATE KEY',
    color: 'red', cx: 720, cy: 1040, w: 290, h: 60, step: 7,
    tooltip: 'The private decapsulation key. Contains the serialised secret vector s, the full ek, its SHA3-256 hash H(ek), and the rejection value z. Required to recover the shared secret from a ciphertext. Must be kept absolutely secret. Sizes: 1632/2400/3168 bytes.',
    definition: 'dk = ByteEncode(s) ‖ ek ‖ H(ek) ‖ z.  |dk| = 1632/2400/3168 bytes for k=2/3/4.  H = SHA3-256.',
    wikiUrl: 'https://en.wikipedia.org/wiki/Key_encapsulation_mechanism',
    specUrl: 'https://doi.org/10.6028/NIST.FIPS.203', specLabel: 'FIPS 203 §5.1 Alg.13',
  },
];

// ── Arrow data ───────────────────────────────────────────────────────────────
const ARROWS: Arrow[] = [
  { from: 'csprng',    to: 'd' },
  { from: 'csprng',    to: 'z' },
  { from: 'd',         to: 'G' },
  { from: 'G',         to: 'rho',      label: 'bytes 0–31' },
  { from: 'G',         to: 'sigma',    label: 'bytes 32–63' },
  { from: 'rho',       to: 'shake128' },
  { from: 'sigma',     to: 'prf' },
  { from: 'shake128',  to: 'sampleNTT' },
  { from: 'prf',       to: 'cbd' },
  { from: 'sampleNTT', to: 'matA' },
  { from: 'cbd',       to: 's' },
  { from: 'cbd',       to: 'e' },
  { from: 'matA',      to: 'ntts', label: 'Â (NTT domain)' },
  { from: 's',         to: 'ntts' },
  { from: 'ntts',      to: 't' },
  { from: 'e',         to: 't' },
  { from: 't',         to: 'ek' },
  { from: 's',         to: 'dk', dashed: true },
  { from: 'z',         to: 'dk', dashed: true },
  { from: 'ek',        to: 'dk', dashed: true },
];

// ── Step divider labels ──────────────────────────────────────────────────────
const STEP_LABELS = [
  { y: 22,   text: 'ENTROPY INPUT' },
  { y: 132,  text: 'SEED SPLITTING' },
  { y: 248,  text: 'KEY EXPANSION  (G = SHA3-512)' },
  { y: 356,  text: 'MATRIX & NOISE GENERATION' },
  { y: 672,  text: 'NTT DOMAIN MULTIPLICATION' },
  { y: 896,  text: 'LWE PUBLIC VECTOR  t' },
  { y: 1002, text: 'KEY SERIALISATION' },
];

// ── Helpers ──────────────────────────────────────────────────────────────────
const byId = (id: string) => NODES.find(n => n.id === id)!;

function arrowPath(a: Node, b: Node): string {
  const x1 = a.cx, y1 = a.cy + a.h / 2;
  const x2 = b.cx, y2 = b.cy - b.h / 2;
  const my = (y1 + y2) / 2;
  return `M ${x1} ${y1} C ${x1} ${my}, ${x2} ${my}, ${x2} ${y2}`;
}

const TT_W = 360, TT_PAD = 14;

function wrapText(text: string, maxChars: number): string[] {
  const words = text.split(' ');
  const lines: string[] = [];
  let line = '';
  for (const w of words) {
    if ((line + ' ' + w).trim().length > maxChars) { lines.push(line.trim()); line = w; }
    else line = (line + ' ' + w).trim();
  }
  if (line) lines.push(line.trim());
  return lines;
}

// ── Summary section content ──────────────────────────────────────────────────
const SUMMARY_STEPS = [
  {
    n: '1',
    title: 'CSPRNG: the only source of randomness',
    text: 'The process begins with a Cryptographically Secure Pseudo-Random Number Generator (CSPRNG), which draws entropy from the operating system or hardware. It produces two independent 32-byte values: d and z.',
  },
  {
    n: '2',
    title: 'G = SHA3-512(d): seed expansion',
    text: 'd is fed into SHA3-512, which outputs 64 bytes. The first 32 bytes become ρ (rho), the public matrix seed and the last 32 bytes become σ (sigma) the private noise seed.',
  },
  {
    n: '3',
    title: 'Matrix Â  public, lives in NTT domain',
    text: 'ρ is expanded by SHAKE-128 (an extendable-output function) through SampleNTT. This builds the k×k matrix Â of polynomials directly in the NTT domain. For ML-KEM-512 (k=2), Â is a 2×2 matrix where each entry is a polynomial of 256 coefficients mod q=3329.',
  },
  {
    n: '4',
    title: 'Secret vector s and error vector e',
    text: 'σ is fed into SHAKE-256 (PRF). The outputs drive a Centered Binomial Distribution (CBD) sampler which produces small-coefficient polynomials: s (the secret vector, k polynomials) and e (the error vector, k polynomials). Both have coefficients in {−η, …, +η}.',
  },
  {
    n: '5',
    title: 'NTT transform   O(n²) → O(n log n)',
    text: 'To multiply Â and s efficiently, s is converted to the NTT domain: ŝ = NTT(s). The matrix-vector product Â ⊗ ŝ is then computed as a pointwise (Hadamard) product, far cheaper than standard polynomial multiplication.',
  },
  {
    n: '6',
    title: 'Public vector t  the LWE instance',
    text: 'The inverse NTT (INTT) is applied to Â ⊗ ŝ to return to standard domain. The error vector e is then added: t = INTT(Â ⊗ ŝ) + e. This is the Module-LWE hardness instance, given Â and t, recovering s is computationally infeasible because of the small noise e.',
  },
  {
    n: '7',
    title: 'Key assembly',
    text: 'The encapsulation key ek = ByteEncode(t) ‖ ρ is the public key  share it freely. The decapsulation key dk = ByteEncode(s) ‖ ek ‖ H(ek) ‖ z bundles the secret s, the full ek, its hash, and the rejection value z , keep it private.',
  },
];

// ── Component ────────────────────────────────────────────────────────────────
export function FlowDiagram() {
  const [hoverId,  setHoverId]  = useState<string | null>(null);
  const [pinnedId, setPinnedId] = useState<string | null>(null);

  const activeId   = pinnedId ?? hoverId;
  const activeNode = activeId ? byId(activeId) : null;

  // theme-aware colours
  const ttBg   = cssVar('--svg-tt-bg',   '#ffffff');
  const ttText = cssVar('--svg-tt-text', '#000000');
  const ttDef  = cssVar('--svg-tt-def',  '#1a1a2e');
  const arrowCol  = cssVar('--text-muted',  '#6b7280');
  const stepColor = cssVar('--text-muted',  '#6b7280');

  // ── Tooltip renderer ────────────────────────────────────────────────────
  const renderTooltip = (node: Node, isPinned: boolean) => {
    const VW = 960;
    const descLines = wrapText(node.tooltip,    44);
    const defLines  = wrapText(node.definition, 44);
    const hdrLines  = wrapText(node.label,       36);

    const closeH = isPinned ? 24 : 0;
    const hdrH   = hdrLines.length  * 20 + 12;
    const descH  = descLines.length * 17 + 6;
    const defH   = defLines.length  * 16 + 4;
    const LINKS_H = 24;
    const boxH = closeH + hdrH + descH + 8 + 14 + 16 + defH + 8 + LINKS_H + 12;

    let tx = node.cx - TT_W / 2;
    if (tx < 6) tx = 6;
    if (tx + TT_W > VW - 6) tx = VW - TT_W - 6;

    const spaceAbove = node.cy - node.h / 2 - 18;
    const above = spaceAbove >= boxH;
    const ty = above ? node.cy - node.h / 2 - boxH - 18 : node.cy + node.h / 2 + 18;

    const caretX    = Math.min(Math.max(node.cx, tx + 14), tx + TT_W - 14);
    const caretBase = above ? ty + boxH : ty;
    const caretTip  = above ? node.cy - node.h / 2 - 6 : node.cy + node.h / 2 + 6;

    let curY = ty + closeH;
    const closeBtnY = ty + 16;
    curY += hdrH + 6;
    const descStartY = curY;  curY += descH + 8;
    const div1Y      = curY;  curY += 14;
    const defLabelY  = curY + 13; curY += 18;
    const defStartY  = curY;  curY += defH + 8;
    const div2Y      = curY;  curY += 8;
    const linksY     = curY + 14;

    const stroke = NODE_STROKE[node.color];

    return (
      <g key="tt" style={{ pointerEvents: 'none' }}>
        {/* shadow */}
        <rect x={tx+4} y={ty+4} width={TT_W} height={boxH} rx="10" fill="rgba(0,0,0,0.18)" />
        {/* background */}
        <rect x={tx} y={ty} width={TT_W} height={boxH} rx="10"
          fill={ttBg} stroke={stroke} strokeWidth="1.5" />
        {/* close button */}
        {isPinned && (
          <g style={{ pointerEvents: 'all' }}>
            <rect x={tx + TT_W - 28} y={ty + 6} width={22} height={22} rx="5"
              fill="rgba(0,0,0,0.06)" style={{ cursor: 'pointer' }}
              onClick={() => setPinnedId(null)} />
            <text x={tx + TT_W - 17} y={closeBtnY}
              textAnchor="middle" fontSize="13" fill={stroke}
              fontFamily="'Inter',system-ui" fontWeight="700"
              style={{ cursor: 'pointer', userSelect: 'none' }}
              onClick={() => setPinnedId(null)}>✕</text>
          </g>
        )}
        {/* header band */}
        <rect x={tx} y={ty + closeH} width={TT_W} height={hdrH} rx="10" fill={NODE_FILL[node.color]} />
        <rect x={tx} y={ty + closeH + hdrH - 10} width={TT_W} height={10} fill={NODE_FILL[node.color]} />
        {/* header text */}
        {hdrLines.map((ln, i) => (
          <text key={i} x={tx + TT_PAD} y={ty + closeH + 18 + i * 20}
            fontSize="13" fontWeight="800" fill={NODE_LABEL[node.color]}
            fontFamily="'JetBrains Mono','Fira Code',monospace">{ln}</text>
        ))}
        {/* description */}
        {descLines.map((ln, i) => (
          <text key={i} x={tx + TT_PAD} y={descStartY + i * 17 + 14}
            fontSize="12" fill={ttText} fontFamily="'Inter',system-ui">{ln}</text>
        ))}
        {/* divider */}
        <line x1={tx + TT_PAD} y1={div1Y} x2={tx + TT_W - TT_PAD} y2={div1Y}
          stroke={stroke} strokeWidth="0.8" strokeOpacity="0.3" />
        {/* DEFINITION label */}
        <text x={tx + TT_PAD} y={defLabelY}
          fontSize="10" fontWeight="700" fill={stroke}
          fontFamily="'Inter',system-ui" letterSpacing="1">FORMAL DEFINITION</text>
        {/* definition lines */}
        {defLines.map((ln, i) => (
          <text key={i} x={tx + TT_PAD} y={defStartY + i * 16 + 13}
            fontSize="11" fill={ttDef} fontFamily="'JetBrains Mono','Fira Code',monospace">{ln}</text>
        ))}
        {/* divider 2 */}
        <line x1={tx + TT_PAD} y1={div2Y} x2={tx + TT_W - TT_PAD} y2={div2Y}
          stroke={stroke} strokeWidth="0.8" strokeOpacity="0.3" />
        {/* links */}
        <g style={{ pointerEvents: 'all' }}>
          <text x={tx + TT_PAD} y={linksY}
            fontSize="11" fill="#1d4ed8" fontFamily="'Inter',system-ui"
            textDecoration="underline" style={{ cursor: 'pointer' }}
            onClick={() => window.open(node.wikiUrl, '_blank', 'noopener,noreferrer')}>
            Wikipedia ↗
          </text>
        </g>
        <g style={{ pointerEvents: 'all' }}>
          <text x={tx + TT_PAD + 100} y={linksY}
            fontSize="11" fill="#1d4ed8" fontFamily="'Inter',system-ui"
            textDecoration="underline" style={{ cursor: 'pointer' }}
            onClick={() => window.open(node.specUrl, '_blank', 'noopener,noreferrer')}>
            {node.specLabel} ↗
          </text>
        </g>
        {/* caret */}
        <polygon
          points={`${caretX - 8},${caretBase} ${caretX + 8},${caretBase} ${caretX},${caretTip}`}
          fill={ttBg} stroke={stroke} strokeWidth="1.5" />
      </g>
    );
  };

  // ── JSX ──────────────────────────────────────────────────────────────────
  return (
    <section id="flow-diagram" className="flow-section" aria-labelledby="flow-heading">

      {/* ── Header ── */}
      <div className="flow-section__header">
        <div className="intro__badge">
          <a href="https://doi.org/10.6028/NIST.FIPS.203" target="_blank"
            rel="noopener noreferrer" className="intro__badge-link">FIPS&nbsp;203</a>
          {'\u00a0·\u00a0ML-KEM Key Generation'}
        </div>
        <h2 id="flow-heading" className="flow-section__title">
          ML-KEM Key Generation
        </h2>
        <p className="flow-section__lead">
          Hover any block to see its description · Click to pin · Click ✕ to dismiss
        </p>
      </div>

      {/* ── Plain-language summary ── */}
      <div className="flow-summary" aria-label="plain-language key generation summary">
        <h3 className="flow-summary__title">How it works</h3>
        <ol className="flow-summary__list">
          {SUMMARY_STEPS.map(s => (
            <li key={s.n} className="flow-summary__item">
              <span className="flow-summary__num">{s.n}</span>
              <div>
                <strong className="flow-summary__step-title">{s.title}</strong>
                <p className="flow-summary__step-text">{s.text}</p>
              </div>
            </li>
          ))}
        </ol>
      </div>

      {/* ── SVG diagram ── */}
      <div className="flow-svg-wrap">
        <svg viewBox="0 0 960 1120" className="flow-svg"
          role="img" aria-label="ML-KEM key generation flow diagram">
          <defs>
            <marker id="arr" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
              <polygon points="0 0,10 3.5,0 7" fill={arrowCol} />
            </marker>
            <marker id="arr-on" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
              <polygon points="0 0,10 3.5,0 7" fill="#1d4ed8" />
            </marker>
          </defs>

          {/* Step divider labels */}
          {STEP_LABELS.map(({ y, text }) => (
            <g key={y}>
              <line x1="40" y1={y} x2="920" y2={y}
                stroke={stepColor} strokeWidth="0.8" strokeDasharray="4,4" strokeOpacity="0.5" />
              <text x="480" y={y - 4} textAnchor="middle"
                fontSize="10" fill={stepColor} fontWeight="700"
                fontFamily="'Inter',system-ui" letterSpacing="1.2">{text}</text>
            </g>
          ))}

          {/* Arrows — rendered behind nodes */}
          {ARROWS.map((arr, i) => {
            const a = byId(arr.from), b = byId(arr.to);
            const isOn = activeId === arr.from || activeId === arr.to;
            const mx = (a.cx + b.cx) / 2;
            const my = (a.cy + a.h / 2 + b.cy - b.h / 2) / 2;
            return (
              <g key={i}>
                <path d={arrowPath(a, b)} fill="none"
                  stroke={isOn ? '#1d4ed8' : arrowCol}
                  strokeWidth={isOn ? 2.5 : 1.8}
                  strokeDasharray={arr.dashed ? '6,4' : undefined}
                  markerEnd={isOn ? 'url(#arr-on)' : 'url(#arr)'}
                  style={{ transition: 'stroke 0.15s, stroke-width 0.15s' }} />
                {arr.label && (
                  <text x={mx} y={my - 4} textAnchor="middle"
                    fontSize="10" fill={isOn ? '#1d4ed8' : arrowCol}
                    fontWeight="600" fontFamily="'Inter',system-ui"
                    style={{ transition: 'fill 0.15s' }}>{arr.label}</text>
                )}
              </g>
            );
          })}

          {/* Nodes */}
          {NODES.map(node => {
            const isOn  = activeId === node.id;
            const x = node.cx - node.w / 2, y = node.cy - node.h / 2;
            const stroke = NODE_STROKE[node.color];
            const fill   = node.color === 'neutral' ? (isOn ? '#eff6ff' : '#ffffff') : NODE_FILL[node.color];

            // Split label into up to 2 lines so it fits inside the node box
            const maxLabelChars = Math.floor(node.w / 7.5);
            const labelLines: string[] = [];
            if (node.label.length <= maxLabelChars) {
              labelLines.push(node.label);
            } else {
              // break at last space before limit
              const breakAt = node.label.lastIndexOf(' ', maxLabelChars);
              if (breakAt > 0) {
                labelLines.push(node.label.slice(0, breakAt));
                labelLines.push(node.label.slice(breakAt + 1));
              } else {
                labelLines.push(node.label.slice(0, maxLabelChars));
                labelLines.push(node.label.slice(maxLabelChars));
              }
            }

            // Split sub into up to 2 lines
            const maxSubChars = Math.floor(node.w / 6);
            const subLines: string[] = [];
            if (node.sub.length <= maxSubChars) {
              subLines.push(node.sub);
            } else {
              const bp = node.sub.lastIndexOf(' ', maxSubChars);
              subLines.push(bp > 0 ? node.sub.slice(0, bp) : node.sub.slice(0, maxSubChars));
              subLines.push(bp > 0 ? node.sub.slice(bp + 1) : node.sub.slice(maxSubChars));
            }

            const totalLines = labelLines.length + subLines.length;
            const lineH = 14;
            const blockH = totalLines * lineH;
            const startY = node.cy - blockH / 2 + lineH * 0.75;

            return (
              <g key={node.id}
                style={{ cursor: 'pointer' }}
                onMouseEnter={() => { if (!pinnedId) setHoverId(node.id); }}
                onMouseLeave={() => { if (!pinnedId) setHoverId(null); }}
                onClick={() => { setHoverId(null); setPinnedId(p => p === node.id ? null : node.id); }}
                role="button" tabIndex={0}
                aria-pressed={pinnedId === node.id}
                aria-label={node.label}
                onKeyDown={ev => {
                  if (ev.key === 'Enter') { setHoverId(null); setPinnedId(p => p === node.id ? null : node.id); }
                }}>
                <rect x={x} y={y} width={node.w} height={node.h} rx="8"
                  fill={fill}
                  stroke={stroke}
                  strokeWidth={isOn ? 2.5 : 1.5}
                  style={{
                    transition: 'all 0.15s',
                    filter: isOn ? `drop-shadow(0 2px 8px rgba(0,0,0,0.2))` : 'none',
                  }} />
                {/* label lines */}
                {labelLines.map((ln, i) => (
                  <text key={`l${i}`} x={node.cx} y={startY + i * lineH}
                    textAnchor="middle" fontSize="12" fontWeight="800"
                    fill={NODE_LABEL[node.color]}
                    fontFamily="'JetBrains Mono','Fira Code',monospace">{ln}</text>
                ))}
                {/* sub-label lines */}
                {subLines.map((ln, i) => (
                  <text key={`s${i}`} x={node.cx} y={startY + labelLines.length * lineH + i * (lineH - 1)}
                    textAnchor="middle" fontSize="9" fill={NODE_SUB[node.color]}
                    fontFamily="'Inter',system-ui">{ln}</text>
                ))}
                {/* step number badge — black circle, white digit, top-left corner of node */}
                <circle
                  cx={x + 13}
                  cy={y + 13}
                  r={10}
                  fill="#1a1a1a"
                  style={{ pointerEvents: 'none' }}
                />
                <text
                  x={x + 13}
                  y={y + 17}
                  textAnchor="middle"
                  fontSize="10"
                  fontWeight="800"
                  fill="#ffffff"
                  fontFamily="'Inter',system-ui"
                  style={{ pointerEvents: 'none', userSelect: 'none' }}
                >{node.step}</text>
              </g>
            );
          })}

          {/* Tooltip on top */}
          {activeNode && renderTooltip(activeNode, pinnedId === activeNode.id)}
        </svg>
      </div>

    </section>
  );
}
