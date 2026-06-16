// Package models defines the JSON message protocol exchanged between the
// frontend and the WebSocket backend over the /ws endpoint.
package models

// ── Inbound messages (frontend → backend) ─────────────────────────────────────

// InboundMessage is the envelope for every message sent by the frontend.
// The Type field determines which payload field is populated.
type InboundMessage struct {
	// Type is the message discriminator. Valid values:
	//   "select_flavor"  — choose an ML-KEM security level
	//   "step_next"      — advance to the next key-generation step
	Type string `json:"type"`

	// Flavor is populated when Type == "select_flavor".
	// Valid values: "512", "768", "1024".
	Flavor string `json:"flavor,omitempty"`

	// Step is populated when Type == "step_next".
	// Valid values: "generate_rho_sigma", "generate_matrix_A",
	//               "generate_vectors", "compute_t", "send_public_key".
	Step string `json:"step,omitempty"`
}

// ── Outbound messages (backend → frontend) ────────────────────────────────────

// OutboundMessage is the envelope for every message sent by the backend.
type OutboundMessage struct {
	// Type identifies the message kind. Matches one of the OutboundType* constants.
	Type string `json:"type"`

	// Payload carries the message body. Exactly one of the Payload* fields below
	// is non-nil for a given Type; the rest are omitted from JSON.
	Payload interface{} `json:"payload"`
}

// ── Payload types ──────────────────────────────────────────────────────────────

// ParamsPayload is sent in response to "select_flavor".
type ParamsPayload struct {
	Flavor string `json:"flavor"`
	N      int    `json:"n"`
	Q      int    `json:"q"`
	K      int    `json:"k"`
	Eta1   int    `json:"eta1"`
	Eta2   int    `json:"eta2"`
	Du     int    `json:"du"`
	Dv     int    `json:"dv"`
	PkSize int    `json:"pk_size"`
	SkSize int    `json:"sk_size"`
	CtSize int    `json:"ct_size"`
}

// RhoSigmaPayload is sent after the G(seed) step.
type RhoSigmaPayload struct {
	// Seed is the 32-byte input seed encoded as a hex string.
	Seed string `json:"seed"`
	// Rho is the 32-byte public matrix seed, hex-encoded.
	Rho string `json:"rho"`
	// Sigma is the 32-byte secret/noise seed, hex-encoded.
	Sigma string `json:"sigma"`
}

// MatrixAPayload is sent after GenerateMatrixA.
// Only the first K×K entries are meaningful; the rest are zeroed.
type MatrixAPayload struct {
	// K is the matrix dimension used.
	K int `json:"k"`
	// A holds the NTT-domain matrix coefficients as a 4×4 array of 256-element slices.
	A [4][4][]int32 `json:"a"`
}

// ByteStreamPayload is sent to visualise the raw PRF output bytes before CBD decoding.
type ByteStreamPayload struct {
	// Label identifies which byte stream this is (e.g. "s_0", "e_2").
	Label string `json:"label"`
	// Bytes is the raw PRF output, hex-encoded.
	Bytes string `json:"bytes"`
}

// VectorsPayload is sent after GenerateSecretAndError.
type VectorsPayload struct {
	K int       `json:"k"`
	S [4][]int32 `json:"s"`
	E [4][]int32 `json:"e"`
}

// TComputedPayload is sent after the t = A·s + e computation.
type TComputedPayload struct {
	K int       `json:"k"`
	T [4][]int32 `json:"t"`
}

// PublicKeyPayload is sent after public key encoding.
type PublicKeyPayload struct {
	// PublicKey is the encoded pk bytes, hex-encoded.
	PublicKey string `json:"public_key"`
	// PublicKeySize is the byte length of the encoded public key.
	PublicKeySize int `json:"public_key_size"`
}

// PrivateKeyPayload is sent alongside the public key.
type PrivateKeyPayload struct {
	// PrivateKey is the encoded sk bytes, hex-encoded.
	PrivateKey string `json:"private_key"`
	// PrivateKeySize is the byte length of the encoded private key.
	PrivateKeySize int `json:"private_key_size"`
}

// ErrorPayload wraps a human-readable error message for the frontend.
type ErrorPayload struct {
	Message string `json:"message"`
}

// ── Outbound message type constants ───────────────────────────────────────────

// Outbound type discriminator strings sent in OutboundMessage.Type.
const (
	TypeParams        = "params"
	TypeRhoSigma      = "rho_sigma"
	TypeMatrixA       = "matrix_A"
	TypeByteStream    = "byte_stream"
	TypeVectors       = "vectors"
	TypeTComputed     = "t_computed"
	TypePublicKeySent = "public_key_sent"
	TypePublicKeyRecv = "public_key_recv"
	TypeError         = "error"
)

// ── Inbound step name constants ────────────────────────────────────────────────

// Step name strings used in InboundMessage.Step.
const (
	StepGenerateRhoSigma = "generate_rho_sigma"
	StepGenerateMatrixA  = "generate_matrix_A"
	StepGenerateVectors  = "generate_vectors"
	StepComputeT         = "compute_t"
	StepSendPublicKey    = "send_public_key"
)

// ── ML-KEM flavor constants (mirrored from mlkem package for protocol use) ─────

// Supported ML-KEM flavor strings used in InboundMessage.Flavor.
const (
	Flavor512  = "512"
	Flavor768  = "768"
	Flavor1024 = "1024"
)
