// Package models defines the JSON message protocol exchanged between the
// frontend and the WebSocket backend over the /ws endpoint.
package models

// ── Inbound messages (frontend → backend) ─────────────────────────────────────

// InboundMessage is the envelope for every message sent by the frontend.
type InboundMessage struct {
	// Type is the message discriminator.
	//   "select_flavor"  — choose an ML-KEM security level
	//   "step_next"      — advance to the next key-generation step
	//   "send_message"   — trigger encapsulation + decapsulation
	//   "reset"          — clear all state, return to flavor selection
	Type string `json:"type"`

	// Flavor is populated when Type == "select_flavor".
	Flavor string `json:"flavor,omitempty"`

	// Step is populated when Type == "step_next".
	Step string `json:"step,omitempty"`
}

// ── Outbound messages (backend → frontend) ────────────────────────────────────

// OutboundMessage is the envelope for every message sent by the backend.
type OutboundMessage struct {
	Type    string      `json:"type"`
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
	Seed  string `json:"seed"`  // hex-encoded 32 bytes
	Rho   string `json:"rho"`   // hex-encoded 32 bytes
	Sigma string `json:"sigma"` // hex-encoded 32 bytes
}

// MatrixAPayload is sent after GenerateMatrixA.
type MatrixAPayload struct {
	K int          `json:"k"`
	A [4][4][]int32 `json:"a"`
}

// ByteStreamPayload is sent to visualise raw PRF output bytes.
type ByteStreamPayload struct {
	Label string `json:"label"`
	Bytes string `json:"bytes"` // hex-encoded
}

// VectorsPayload is sent after GenerateSecretAndError.
type VectorsPayload struct {
	K int        `json:"k"`
	S [4][]int32 `json:"s"`
	E [4][]int32 `json:"e"`
}

// TComputedPayload is sent after t = A·s + e.
type TComputedPayload struct {
	K int        `json:"k"`
	T [4][]int32 `json:"t"`
}

// PublicKeyPayload is sent after public key encoding.
type PublicKeyPayload struct {
	PublicKey     string `json:"public_key"`      // hex-encoded
	PublicKeySize int    `json:"public_key_size"` // bytes
}

// PrivateKeyPayload is sent alongside the public key.
type PrivateKeyPayload struct {
	PrivateKey     string `json:"private_key"`      // hex-encoded
	PrivateKeySize int    `json:"private_key_size"` // bytes
}

// EncryptResultPayload is sent after Encapsulate completes.
type EncryptResultPayload struct {
	Ciphertext     string `json:"ciphertext"`      // hex-encoded
	CiphertextSize int    `json:"ciphertext_size"` // bytes
	SharedSecret   string `json:"shared_secret"`   // hex-encoded
	Message        string `json:"message"`         // hex-encoded random plaintext
}

// DecryptResultPayload is sent after Decapsulate completes.
type DecryptResultPayload struct {
	SharedSecret string `json:"shared_secret"` // hex-encoded
	Match        bool   `json:"match"`         // true iff encaps and decaps secrets agree
}

// ErrorPayload wraps a human-readable error message for the frontend.
type ErrorPayload struct {
	Message string `json:"message"`
}

// ── Outbound message type constants ───────────────────────────────────────────

const (
	TypeParams        = "params"
	TypeRhoSigma      = "rho_sigma"
	TypeMatrixA       = "matrix_A"
	TypeByteStream    = "byte_stream"
	TypeVectors       = "vectors"
	TypeTComputed     = "t_computed"
	TypePublicKeySent = "public_key_sent"
	TypePublicKeyRecv = "public_key_recv"
	TypeEncryptResult = "encrypt_result"
	TypeDecryptResult = "decrypt_result"
	TypeReset         = "reset"
	TypeError         = "error"
)

// ── Inbound message type constants ────────────────────────────────────────────

const (
	MsgSelectFlavor = "select_flavor"
	MsgStepNext     = "step_next"
	MsgSendMessage  = "send_message"
	MsgReset        = "reset"
)

// ── Inbound step name constants ────────────────────────────────────────────────

const (
	StepGenerateRhoSigma = "generate_rho_sigma"
	StepGenerateMatrixA  = "generate_matrix_A"
	StepGenerateVectors  = "generate_vectors"
	StepComputeT         = "compute_t"
	StepSendPublicKey    = "send_public_key"
)

// ── ML-KEM flavor constants ────────────────────────────────────────────────────

const (
	Flavor512  = "512"
	Flavor768  = "768"
	Flavor1024 = "1024"
)
