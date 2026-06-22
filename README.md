# ML-KEM Web Application

[![CI](https://github.com/muhammadalihussnain/mlkem-webapp/actions/workflows/ci.yml/badge.svg)](https://github.com/muhammadalihussnain/mlkem-webapp/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/muhammadalihussnain/mlkem-webapp/graph/badge.svg)](https://codecov.io/gh/muhammadalihussnain/mlkem-webapp)

## ML-KEM (CRYSTALS-Kyber) Post-Quantum Cryptography Web Application

This application implements ML-KEM (formerly CRYSTALS-Kyber) as specified in FIPS 203.

## Project Structure
.
├── backend/ # Go backend with ML-KEM implementation
├── frontend/ # React TypeScript frontend
├── .github/ # GitHub Actions CI/CD
└── Makefile # Build and test automation


## Current Status (Day 1)

✅ Repository setup  
✅ CI pipeline configured  
✅ ML-KEM parameters defined (FIPS 203 compliant)  
✅ 100% test coverage for parameters module  

## Interactive Demo

The web app walks you through the full ML-KEM key generation and key exchange process step by step. Open the app in your browser and follow these steps:

### 1 · Pick a Security Level

Use the **Security Level** dropdown to choose one of the three FIPS 203 variants:

| Variant | Security | Encapsulation Key | Decapsulation Key |
|---|---|---|---|
| ML-KEM-512 | Level 1 (≈ AES-128) | 800 bytes | 1632 bytes |
| ML-KEM-768 | Level 3 (≈ AES-192) | 1184 bytes | 2400 bytes |
| ML-KEM-1024 | Level 5 (≈ AES-256) | 1568 bytes | 3168 bytes |

The **Parameters** panel on the right updates immediately to show the exact values for `n`, `q`, `k`, `η₁`, `η₂`, `dᵤ`, and `dᵥ`.

> You must be connected (🟢 Connected in the header) before selecting a level. If you see 🔴 Disconnected, make sure the backend is running.

---

### 2 · Key Generation (Steps 1–7)

The flowchart above the demo and the interactive steps below map to the same 7 stages. Work through them in order — each button unlocks only after the previous step completes.

**Step 1 — ENTROPY INPUT (CSPRNG)**
The process starts with a Cryptographically Secure Pseudo-Random Number Generator (CSPRNG) drawing entropy from the OS or hardware. It produces two independent 32-byte values: `d` (primary seed) and `z` (implicit-rejection secret stored in `dk`). This is the only true source of randomness — everything after is deterministic.

**Step 2 — SEED SPLITTING (d → G = SHA3-512)**
`d` is fed into SHA3-512, which outputs 64 bytes. The first 32 bytes become `ρ` (rho) — the public matrix seed — and the last 32 bytes become `σ` (sigma) — the private noise seed. Being a one-way function, you cannot recover `d` from `ρ` and `σ`.

**Step 3 — KEY EXPANSION (Matrix Â via SHAKE-128)**
`ρ` is expanded by SHAKE-128 (XOF) through SampleNTT, building the k×k matrix Â of polynomials directly in the NTT domain. For ML-KEM-512 (k=2), Â is a 2×2 matrix where each entry is a 256-coefficient polynomial mod q=3329. This matrix is public and included in `ek`.

**Step 4 — MATRIX & NOISE GENERATION (s and e via SHAKE-256 + CBD)**
`σ` is fed into SHAKE-256 (PRF). The outputs drive a Centered Binomial Distribution (CBD) sampler producing small-coefficient polynomials: `s` (the secret vector, k polynomials) and `e` (the error vector, k polynomials). Both have coefficients in {−η, …, +η}. Neither leaves the server.

**Step 5 — NTT DOMAIN MULTIPLICATION (ŝ = NTT(s))**
To multiply Â and s efficiently, s is converted to the NTT domain: `ŝ = NTT(s)`. The matrix-vector product Â ⊗ ŝ is then computed as a pointwise (Hadamard) product — reducing complexity from O(n²) to O(n log n).

**Step 6 — LWE PUBLIC VECTOR t (t = INTT(Â ⊗ ŝ) + e)**
The inverse NTT (INTT) returns Â ⊗ ŝ to the standard domain, then the small error vector `e` is added: `t = INTT(Â ⊗ ŝ) + e`. This is the Module-LWE hardness instance — given Â and t, recovering s is computationally infeasible because of the noise `e`.

**Step 7 — KEY SERIALISATION (ek and dk)**
The encapsulation key `ek = ByteEncode(t) ‖ ρ` is the public key — share it freely. The decapsulation key `dk = ByteEncode(s) ‖ ek ‖ H(ek) ‖ z` bundles the secret `s`, the full `ek`, its SHA3-256 hash, and the rejection value `z` — keep it private.

> Hover over any node in the flowchart or any step button to see a tooltip with a precise technical description and a link to the relevant FIPS 203 section.

---

### 3 · Key Exchange

Once the public key is ready, the **Key Exchange** section unlocks.

Click **Run Key Exchange** (or equivalent button) to:
1. **Encapsulate** — the frontend sends a request; the backend acts as the encapsulator and generates a random ciphertext + shared secret using `ek`.
2. **Decapsulate** — the backend uses `dk` to recover the shared secret from the ciphertext.
3. The UI shows both the ciphertext (hex) and confirms whether the shared secrets match (`✓ match`).

---

### 4 · Event Log

The **Event Log** at the bottom records every WebSocket message exchanged between the browser and the backend in real time — useful for seeing the raw protocol payloads.

---

### Reset

Click the **Reset** button in the header at any time to clear all state and start over from Step 1. The WebSocket connection is preserved.

---

## Running Locally

```bash
# Start both backend and frontend together
docker-compose up

# Or run them separately:

# Backend (Go)
cd backend
go run ./cmd/server

# Frontend (Node)
cd frontend
npm install
npm run dev
```

```bash
# Run all tests
make test

# Check coverage
make coverage
```
