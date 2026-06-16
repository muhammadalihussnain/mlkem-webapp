//go:build integration

// Package integration contains end-to-end tests that spin up a real HTTP server
// and drive the full ML-KEM key-generation and encapsulation flow over WebSocket.
//
// Run with:
//
//	go test ./integration/... -tags=integration -v
package integration

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"ntt_verification/models"
	"ntt_verification/server"
)

// ── Constants ──────────────────────────────────────────────────────────────────

// recvTimeout is the maximum time a client will wait for a single message.
const recvTimeout = 10 * time.Second

// ── Test helpers ───────────────────────────────────────────────────────────────

// testServerURL spins up a real httptest.Server and returns its WebSocket base URL.
func testServerURL(t *testing.T) (string, func()) {
	t.Helper()
	srv := httptest.NewServer(server.NewHandler(server.Config{}))
	wsBase := "ws" + strings.TrimPrefix(srv.URL, "http")
	return wsBase, srv.Close
}

// dial opens a WebSocket connection to the test server.
func dial(t *testing.T, wsBase string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(wsBase+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn
}

// send marshals msg and writes it to conn.
func send(t *testing.T, conn *websocket.Conn, msg models.InboundMessage) {
	t.Helper()
	b, _ := json.Marshal(msg)
	if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
		t.Fatalf("send: %v", err)
	}
}

// recv reads one message and returns it as an OutboundMessage.
func recv(t *testing.T, conn *websocket.Conn) models.OutboundMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(recvTimeout))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	var msg models.OutboundMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return msg
}

// payloadAs round-trips the payload through JSON into dst.
func payloadAs(t *testing.T, msg models.OutboundMessage, dst interface{}) {
	t.Helper()
	b, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(b, dst); err != nil {
		t.Fatalf("payloadAs %T: %v", dst, err)
	}
}

// inbound builds a convenience InboundMessage.
func inbound(msgType, flavor, step string) models.InboundMessage {
	return models.InboundMessage{Type: msgType, Flavor: flavor, Step: step}
}

// ── Integration tests ──────────────────────────────────────────────────────────

// TestTwoClientsBroadcast verifies that messages sent by one client are received
// by all connected clients.  We connect two clients, drive the full key-generation
// flow from client A, and assert client B receives every broadcast message.
func TestTwoClientsBroadcast(t *testing.T) {
	wsBase, close := testServerURL(t)
	defer close()

	clientA := dial(t, wsBase)
	defer clientA.Close()
	clientB := dial(t, wsBase)
	defer clientB.Close()

	// ── select_flavor ──────────────────────────────────────────────────────────
	send(t, clientA, inbound(models.MsgSelectFlavor, models.Flavor512, ""))

	for _, name := range []string{"clientA", "clientB"} {
		var conn *websocket.Conn
		if name == "clientA" {
			conn = clientA
		} else {
			conn = clientB
		}
		msg := recv(t, conn)
		if msg.Type != models.TypeParams {
			t.Fatalf("%s: want %q got %q", name, models.TypeParams, msg.Type)
		}
		var p models.ParamsPayload
		payloadAs(t, msg, &p)
		if p.K != 2 {
			t.Errorf("%s: params k want 2 got %d", name, p.K)
		}
	}

	// ── generate_rho_sigma ────────────────────────────────────────────────────
	send(t, clientA, inbound(models.MsgStepNext, "", models.StepGenerateRhoSigma))
	rhoMsgA := recv(t, clientA)
	rhoMsgB := recv(t, clientB)
	if rhoMsgA.Type != models.TypeRhoSigma || rhoMsgB.Type != models.TypeRhoSigma {
		t.Fatalf("want rho_sigma on both clients")
	}

	// Both clients receive identical rho/sigma.
	var rsA, rsB models.RhoSigmaPayload
	payloadAs(t, rhoMsgA, &rsA)
	payloadAs(t, rhoMsgB, &rsB)
	if rsA.Rho != rsB.Rho {
		t.Errorf("rho mismatch between clients: A=%s B=%s", rsA.Rho, rsB.Rho)
	}

	// ── generate_matrix_A ────────────────────────────────────────────────────
	send(t, clientA, inbound(models.MsgStepNext, "", models.StepGenerateMatrixA))
	matA := recv(t, clientA)
	matB := recv(t, clientB)
	if matA.Type != models.TypeMatrixA || matB.Type != models.TypeMatrixA {
		t.Fatalf("want matrix_A on both clients")
	}

	// ── generate_vectors ──────────────────────────────────────────────────────
	send(t, clientA, inbound(models.MsgStepNext, "", models.StepGenerateVectors))
	vecA := recv(t, clientA)
	vecB := recv(t, clientB)
	if vecA.Type != models.TypeVectors || vecB.Type != models.TypeVectors {
		t.Fatalf("want vectors on both clients")
	}

	// ── compute_t ─────────────────────────────────────────────────────────────
	send(t, clientA, inbound(models.MsgStepNext, "", models.StepComputeT))
	tA := recv(t, clientA)
	tB := recv(t, clientB)
	if tA.Type != models.TypeTComputed || tB.Type != models.TypeTComputed {
		t.Fatalf("want t_computed on both clients")
	}

	// ── send_public_key ───────────────────────────────────────────────────────
	send(t, clientA, inbound(models.MsgStepNext, "", models.StepSendPublicKey))

	// Expect public_key_sent then public_key_recv on both clients.
	for _, conn := range []*websocket.Conn{clientA, clientB} {
		pkSent := recv(t, conn)
		if pkSent.Type != models.TypePublicKeySent {
			t.Fatalf("want public_key_sent got %q", pkSent.Type)
		}
		var pk models.PublicKeyPayload
		payloadAs(t, pkSent, &pk)
		if pk.PublicKeySize != 800 { // ML-KEM-512
			t.Errorf("pk size want 800 got %d", pk.PublicKeySize)
		}

		pkRecv := recv(t, conn)
		if pkRecv.Type != models.TypePublicKeyRecv {
			t.Fatalf("want public_key_recv got %q", pkRecv.Type)
		}
		var sk models.PrivateKeyPayload
		payloadAs(t, pkRecv, &sk)
		if sk.PrivateKeySize != 768 { // ML-KEM-512
			t.Errorf("sk size want 768 got %d", sk.PrivateKeySize)
		}
	}
}

// TestEncapsDecapsBothClients drives the full flow including encapsulation and
// verifies both clients receive matching shared secrets.
func TestEncapsDecapsBothClients(t *testing.T) {
	for _, flavor := range []string{models.Flavor512, models.Flavor768, models.Flavor1024} {
		t.Run(flavor, func(t *testing.T) {
			wsBase, close := testServerURL(t)
			defer close()

			clientA := dial(t, wsBase)
			defer clientA.Close()
			clientB := dial(t, wsBase)
			defer clientB.Close()

			// Drive full key-gen sequence from clientA; consume from both.
			steps := []struct {
				msg      models.InboundMessage
				wantType string
				extra    string // second message type if any
			}{
				{inbound(models.MsgSelectFlavor, flavor, ""), models.TypeParams, ""},
				{inbound(models.MsgStepNext, "", models.StepGenerateRhoSigma), models.TypeRhoSigma, ""},
				{inbound(models.MsgStepNext, "", models.StepGenerateMatrixA), models.TypeMatrixA, ""},
				{inbound(models.MsgStepNext, "", models.StepGenerateVectors), models.TypeVectors, ""},
				{inbound(models.MsgStepNext, "", models.StepComputeT), models.TypeTComputed, ""},
				{inbound(models.MsgStepNext, "", models.StepSendPublicKey), models.TypePublicKeySent, models.TypePublicKeyRecv},
			}

			for _, s := range steps {
				send(t, clientA, s.msg)
				// Both clients must receive primary message.
				for _, conn := range []*websocket.Conn{clientA, clientB} {
					msg := recv(t, conn)
					if msg.Type != s.wantType {
						t.Fatalf("step %q: want %q got %q", s.msg.Step, s.wantType, msg.Type)
					}
				}
				// Consume second message if expected.
				if s.extra != "" {
					for _, conn := range []*websocket.Conn{clientA, clientB} {
						msg := recv(t, conn)
						if msg.Type != s.extra {
							t.Fatalf("step %q extra: want %q got %q", s.msg.Step, s.extra, msg.Type)
						}
					}
				}
			}

			// ── Encapsulate / Decapsulate ──────────────────────────────────────
			send(t, clientA, models.InboundMessage{Type: models.MsgSendMessage})

			// Collect encrypt_result from both clients.
			encA := recv(t, clientA)
			encB := recv(t, clientB)
			if encA.Type != models.TypeEncryptResult || encB.Type != models.TypeEncryptResult {
				t.Fatalf("want encrypt_result on both clients")
			}

			var epA, epB models.EncryptResultPayload
			payloadAs(t, encA, &epA)
			payloadAs(t, encB, &epB)

			// Both clients receive the same ciphertext.
			if epA.Ciphertext != epB.Ciphertext {
				t.Error("ciphertext differs between clients")
			}

			// Collect decrypt_result from both clients.
			decA := recv(t, clientA)
			decB := recv(t, clientB)
			if decA.Type != models.TypeDecryptResult || decB.Type != models.TypeDecryptResult {
				t.Fatalf("want decrypt_result on both clients")
			}

			var dpA, dpB models.DecryptResultPayload
			payloadAs(t, decA, &dpA)
			payloadAs(t, decB, &dpB)

			if !dpA.Match {
				t.Errorf("flavor=%s: shared secrets do not match on clientA", flavor)
			}
			if epA.SharedSecret != dpA.SharedSecret {
				t.Errorf("flavor=%s: encaps/decaps secret mismatch", flavor)
			}
			if dpA.SharedSecret != dpB.SharedSecret {
				t.Error("decrypt shared_secret differs between clients")
			}
		})
	}
}

// TestResetClearsState verifies that a reset message returns the session to
// the initial state and broadcasts a reset event to all connected clients.
func TestResetClearsState(t *testing.T) {
	wsBase, close := testServerURL(t)
	defer close()

	clientA := dial(t, wsBase)
	defer clientA.Close()
	clientB := dial(t, wsBase)
	defer clientB.Close()

	// Select a flavor first.
	send(t, clientA, inbound(models.MsgSelectFlavor, models.Flavor512, ""))
	recv(t, clientA)
	recv(t, clientB)

	// Reset from clientA.
	send(t, clientA, models.InboundMessage{Type: models.MsgReset})

	// Both clients should receive a reset broadcast.
	for _, conn := range []*websocket.Conn{clientA, clientB} {
		msg := recv(t, conn)
		if msg.Type != models.TypeReset {
			t.Fatalf("want %q got %q", models.TypeReset, msg.Type)
		}
	}

	// After reset, step_next without a flavor should return an error.
	send(t, clientA, inbound(models.MsgStepNext, "", models.StepGenerateRhoSigma))
	errMsg := recv(t, clientA)
	if errMsg.Type != models.TypeError {
		t.Fatalf("after reset: want error got %q", errMsg.Type)
	}
}
