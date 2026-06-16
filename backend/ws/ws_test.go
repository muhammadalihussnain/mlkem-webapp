package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"ntt_verification/models"
)

// dialTimeout is the maximum time allowed for the test client to connect.
const dialTimeout = 2 * time.Second

// recvTimeout is the maximum time to wait for a single message from the server.
const recvTimeout = 5 * time.Second

// ── Test helpers ───────────────────────────────────────────────────────────────

// testServer spins up an httptest.Server with a Hub and returns the server and
// a connected WebSocket client.
func testServer(t *testing.T) (*httptest.Server, *websocket.Conn) {
	t.Helper()
	hub := NewHub()
	srv := httptest.NewServer(http.HandlerFunc(Handler(hub)))

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		srv.Close()
		t.Fatalf("dial: %v", err)
	}
	return srv, conn
}

// send marshals v and writes it as a text message.
func send(t *testing.T, conn *websocket.Conn, v interface{}) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal send: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// recv reads one text message and unmarshals it into an OutboundMessage.
func recv(t *testing.T, conn *websocket.Conn) models.OutboundMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(recvTimeout))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var msg models.OutboundMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return msg
}

// payloadAs unmarshals the payload of msg into dst using JSON round-trip.
func payloadAs(t *testing.T, msg models.OutboundMessage, dst interface{}) {
	t.Helper()
	b, err := json.Marshal(msg.Payload)
	if err != nil {
		t.Fatalf("re-marshal payload: %v", err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		t.Fatalf("unmarshal payload into %T: %v", dst, err)
	}
}

// inbound is a convenience constructor.
func inbound(msgType, flavor, step string) models.InboundMessage {
	return models.InboundMessage{Type: msgType, Flavor: flavor, Step: step}
}

// ── State transition tests ─────────────────────────────────────────────────────

// TestSelectFlavor verifies that sending select_flavor broadcasts a params message
// for each supported security level.
func TestSelectFlavor(t *testing.T) {
	for _, flavor := range []string{models.Flavor512, models.Flavor768, models.Flavor1024} {
		flavor := flavor
		t.Run(flavor, func(t *testing.T) {
			srv, conn := testServer(t)
			defer srv.Close()
			defer conn.Close()

			send(t, conn, inbound("select_flavor", flavor, ""))

			msg := recv(t, conn)
			if msg.Type != models.TypeParams {
				t.Fatalf("want type %q, got %q", models.TypeParams, msg.Type)
			}

			var p models.ParamsPayload
			payloadAs(t, msg, &p)

			if p.Flavor != flavor {
				t.Errorf("flavor: want %q got %q", flavor, p.Flavor)
			}
			if p.N != 256 {
				t.Errorf("N: want 256 got %d", p.N)
			}
			if p.Q != 3329 {
				t.Errorf("Q: want 3329 got %d", p.Q)
			}
			if p.K < 2 || p.K > 4 {
				t.Errorf("K: want 2-4 got %d", p.K)
			}
		})
	}
}

// TestStepGenerateRhoSigma verifies the generate_rho_sigma step returns valid hex seeds.
func TestStepGenerateRhoSigma(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("select_flavor", models.Flavor768, ""))
	recv(t, conn) // discard params

	send(t, conn, inbound("step_next", "", models.StepGenerateRhoSigma))

	msg := recv(t, conn)
	if msg.Type != models.TypeRhoSigma {
		t.Fatalf("want type %q, got %q", models.TypeRhoSigma, msg.Type)
	}

	var p models.RhoSigmaPayload
	payloadAs(t, msg, &p)

	// Each value should be a 32-byte hex string (64 hex chars).
	const hexLen32 = 64
	if len(p.Seed) != hexLen32 {
		t.Errorf("seed hex length: want %d got %d", hexLen32, len(p.Seed))
	}
	if len(p.Rho) != hexLen32 {
		t.Errorf("rho hex length: want %d got %d", hexLen32, len(p.Rho))
	}
	if len(p.Sigma) != hexLen32 {
		t.Errorf("sigma hex length: want %d got %d", hexLen32, len(p.Sigma))
	}
	if p.Seed == p.Rho {
		t.Error("seed and rho should differ")
	}
}

// TestStepGenerateMatrixA verifies matrix A is populated for k×k entries.
func TestStepGenerateMatrixA(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("select_flavor", models.Flavor512, ""))
	recv(t, conn)

	send(t, conn, inbound("step_next", "", models.StepGenerateRhoSigma))
	recv(t, conn)

	send(t, conn, inbound("step_next", "", models.StepGenerateMatrixA))
	msg := recv(t, conn)

	if msg.Type != models.TypeMatrixA {
		t.Fatalf("want type %q, got %q", models.TypeMatrixA, msg.Type)
	}

	var p models.MatrixAPayload
	payloadAs(t, msg, &p)

	if p.K != 2 {
		t.Errorf("k: want 2 got %d", p.K)
	}
	// Each populated entry should have N=256 coefficients.
	for i := 0; i < p.K; i++ {
		for j := 0; j < p.K; j++ {
			if len(p.A[i][j]) != 256 {
				t.Errorf("A[%d][%d] length: want 256 got %d", i, j, len(p.A[i][j]))
			}
		}
	}
}

// TestStepGenerateVectors verifies s and e vectors have correct dimensions.
func TestStepGenerateVectors(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("select_flavor", models.Flavor768, ""))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepGenerateRhoSigma))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepGenerateVectors))

	msg := recv(t, conn)
	if msg.Type != models.TypeVectors {
		t.Fatalf("want type %q, got %q", models.TypeVectors, msg.Type)
	}

	var p models.VectorsPayload
	payloadAs(t, msg, &p)

	if p.K != 3 {
		t.Errorf("k: want 3 got %d", p.K)
	}
	for i := 0; i < p.K; i++ {
		if len(p.S[i]) != 256 {
			t.Errorf("s[%d] length: want 256 got %d", i, len(p.S[i]))
		}
		if len(p.E[i]) != 256 {
			t.Errorf("e[%d] length: want 256 got %d", i, len(p.E[i]))
		}
	}
}

// TestStepComputeT verifies the t vector has correct dimensions.
func TestStepComputeT(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("select_flavor", models.Flavor512, ""))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepGenerateRhoSigma))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepGenerateMatrixA))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepGenerateVectors))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepComputeT))

	msg := recv(t, conn)
	if msg.Type != models.TypeTComputed {
		t.Fatalf("want type %q, got %q", models.TypeTComputed, msg.Type)
	}

	var p models.TComputedPayload
	payloadAs(t, msg, &p)

	if p.K != 2 {
		t.Errorf("k: want 2 got %d", p.K)
	}
	for i := 0; i < p.K; i++ {
		if len(p.T[i]) != 256 {
			t.Errorf("t[%d] length: want 256 got %d", i, len(p.T[i]))
		}
	}
}

// TestStepSendPublicKey verifies public and private key messages are broadcast.
func TestStepSendPublicKey(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	// Run the full sequence for ML-KEM-512.
	send(t, conn, inbound("select_flavor", models.Flavor512, ""))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepGenerateRhoSigma))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepGenerateMatrixA))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepGenerateVectors))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepComputeT))
	recv(t, conn)
	send(t, conn, inbound("step_next", "", models.StepSendPublicKey))

	// Expect two messages: public_key_sent then public_key_recv.
	pkSent := recv(t, conn)
	if pkSent.Type != models.TypePublicKeySent {
		t.Fatalf("want %q got %q", models.TypePublicKeySent, pkSent.Type)
	}
	var pk models.PublicKeyPayload
	payloadAs(t, pkSent, &pk)
	// ML-KEM-512 pk = 768 bytes (t) + 32 bytes (rho) = 800 bytes → 1600 hex chars.
	if pk.PublicKeySize != 800 {
		t.Errorf("pk size: want 800 got %d", pk.PublicKeySize)
	}

	pkRecv := recv(t, conn)
	if pkRecv.Type != models.TypePublicKeyRecv {
		t.Fatalf("want %q got %q", models.TypePublicKeyRecv, pkRecv.Type)
	}
	var sk models.PrivateKeyPayload
	payloadAs(t, pkRecv, &sk)
	// ML-KEM-512 sk = 768 bytes.
	if sk.PrivateKeySize != 768 {
		t.Errorf("sk size: want 768 got %d", sk.PrivateKeySize)
	}
}

// TestStepNextWithoutFlavor verifies an error is returned when step_next is
// sent before select_flavor.
func TestStepNextWithoutFlavor(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("step_next", "", models.StepGenerateRhoSigma))

	msg := recv(t, conn)
	if msg.Type != models.TypeError {
		t.Fatalf("want type %q, got %q", models.TypeError, msg.Type)
	}
}

// TestHubClientCount verifies the hub tracks registered clients correctly.
func TestHubClientCount(t *testing.T) {
	hub := NewHub()
	if hub.ClientCount() != 0 {
		t.Fatalf("initial count: want 0 got %d", hub.ClientCount())
	}

	c := &client{send: make(chan []byte, 1)}
	hub.register(c)
	if hub.ClientCount() != 1 {
		t.Fatalf("after register: want 1 got %d", hub.ClientCount())
	}

	hub.unregister(c)
	if hub.ClientCount() != 0 {
		t.Fatalf("after unregister: want 0 got %d", hub.ClientCount())
	}
}

// TestHubBroadcast verifies messages reach registered clients.
func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	c := &client{send: make(chan []byte, 4)}
	hub.register(c)

	msg := []byte(`{"type":"test"}`)
	hub.Broadcast(msg)

	select {
	case got := <-c.send:
		if string(got) != string(msg) {
			t.Errorf("broadcast: want %q got %q", msg, got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("broadcast: timed out waiting for message")
	}
}

// TestHubBroadcastFullBuffer verifies a full send buffer does not block the broadcaster.
func TestHubBroadcastFullBuffer(t *testing.T) {
	hub := NewHub()
	// Buffer size 1 — will be full after the first message.
	c := &client{send: make(chan []byte, 1)}
	hub.register(c)

	// Fill the buffer.
	hub.Broadcast([]byte(`{"type":"a"}`))
	// This second broadcast should not block or panic.
	hub.Broadcast([]byte(`{"type":"b"}`))
}

// TestSelectInvalidFlavor verifies that an invalid flavor returns an error message.
func TestSelectInvalidFlavor(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("select_flavor", "9999", ""))

	msg := recv(t, conn)
	if msg.Type != models.TypeError {
		t.Fatalf("want type %q, got %q", models.TypeError, msg.Type)
	}
}

// TestUnknownMessageType verifies that unknown message types are silently ignored
// (no response is sent, and the next valid message is handled correctly).
func TestUnknownMessageType(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	// Send an unknown type — server should not respond.
	send(t, conn, models.InboundMessage{Type: "unknown_type"})

	// Then send a valid message to confirm the connection is still alive.
	send(t, conn, inbound("select_flavor", models.Flavor512, ""))
	msg := recv(t, conn)
	if msg.Type != models.TypeParams {
		t.Fatalf("after unknown type, want %q got %q", models.TypeParams, msg.Type)
	}
}

// TestStepMatrixAWithoutRhoSigma verifies an error when matrix A is requested
// before rho/sigma have been generated.
func TestStepMatrixAWithoutRhoSigma(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("select_flavor", models.Flavor512, ""))
	recv(t, conn)

	// Skip generate_rho_sigma and jump straight to generate_matrix_A.
	send(t, conn, inbound("step_next", "", models.StepGenerateMatrixA))
	msg := recv(t, conn)
	if msg.Type != models.TypeError {
		t.Fatalf("want %q got %q", models.TypeError, msg.Type)
	}
}

// TestStepVectorsWithoutRhoSigma verifies an error when vectors are requested
// before rho/sigma have been generated.
func TestStepVectorsWithoutRhoSigma(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("select_flavor", models.Flavor512, ""))
	recv(t, conn)

	send(t, conn, inbound("step_next", "", models.StepGenerateVectors))
	msg := recv(t, conn)
	if msg.Type != models.TypeError {
		t.Fatalf("want %q got %q", models.TypeError, msg.Type)
	}
}

// TestStepComputeTWithoutVectors verifies an error when compute_t is called
// before vectors have been generated.
func TestStepComputeTWithoutVectors(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("select_flavor", models.Flavor512, ""))
	recv(t, conn)

	send(t, conn, inbound("step_next", "", models.StepComputeT))
	msg := recv(t, conn)
	if msg.Type != models.TypeError {
		t.Fatalf("want %q got %q", models.TypeError, msg.Type)
	}
}

// TestStepSendPublicKeyWithoutT verifies an error when send_public_key is called
// before compute_t.
func TestStepSendPublicKeyWithoutT(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("select_flavor", models.Flavor512, ""))
	recv(t, conn)

	send(t, conn, inbound("step_next", "", models.StepSendPublicKey))
	msg := recv(t, conn)
	if msg.Type != models.TypeError {
		t.Fatalf("want %q got %q", models.TypeError, msg.Type)
	}
}

// TestInvalidJSON verifies that malformed JSON results in an error response.
func TestInvalidJSON(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	conn.WriteMessage(websocket.TextMessage, []byte(`{not valid json`))
	msg := recv(t, conn)
	if msg.Type != models.TypeError {
		t.Fatalf("want %q got %q", models.TypeError, msg.Type)
	}
}

// TestFullSequenceML-KEM768 runs the complete key-generation sequence for ML-KEM-768.
func TestFullSequenceMLKEM768(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	steps := []struct {
		send models.InboundMessage
		want string
	}{
		{inbound("select_flavor", models.Flavor768, ""), models.TypeParams},
		{inbound("step_next", "", models.StepGenerateRhoSigma), models.TypeRhoSigma},
		{inbound("step_next", "", models.StepGenerateMatrixA), models.TypeMatrixA},
		{inbound("step_next", "", models.StepGenerateVectors), models.TypeVectors},
		{inbound("step_next", "", models.StepComputeT), models.TypeTComputed},
		{inbound("step_next", "", models.StepSendPublicKey), models.TypePublicKeySent},
	}

	for _, s := range steps {
		send(t, conn, s.send)
		msg := recv(t, conn)
		if msg.Type != s.want {
			t.Fatalf("step %q: want type %q got %q", s.send.Step, s.want, msg.Type)
		}
	}

	// Also consume the public_key_recv message.
	extra := recv(t, conn)
	if extra.Type != models.TypePublicKeyRecv {
		t.Fatalf("want %q got %q", models.TypePublicKeyRecv, extra.Type)
	}
}

// TestUnknownStep verifies that an unknown step name is silently ignored.
func TestUnknownStep(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound("select_flavor", models.Flavor512, ""))
	recv(t, conn)

	// Unknown step — server should not respond; confirm connection still works.
	send(t, conn, inbound("step_next", "", "unknown_step"))

	// Send a valid step to confirm the connection is live.
	send(t, conn, inbound("step_next", "", models.StepGenerateRhoSigma))
	msg := recv(t, conn)
	if msg.Type != models.TypeRhoSigma {
		t.Fatalf("want %q got %q", models.TypeRhoSigma, msg.Type)
	}
}

// TestConnectionClose verifies that the server handles a client closing its connection.
func TestConnectionClose(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()

	// Normal interaction first.
	send(t, conn, inbound("select_flavor", models.Flavor512, ""))
	recv(t, conn)

	// Close the connection gracefully — this exercises readPump's exit path
	// and triggers writePump to drain and return.
	conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "done"))
	conn.Close()

	// Give the server a moment to clean up.
	time.Sleep(50 * time.Millisecond)
}

// TestHubUnregisterIdempotent verifies that unregistering the same client twice
// does not panic or corrupt state.
func TestHubUnregisterIdempotent(t *testing.T) {
	hub := NewHub()
	c := &client{send: make(chan []byte, 1)}
	hub.register(c)
	hub.unregister(c)
	hub.unregister(c) // second call must be a no-op
	if hub.ClientCount() != 0 {
		t.Errorf("count after double-unregister: want 0 got %d", hub.ClientCount())
	}
}

// runFullKeyGen runs the complete key-generation sequence on a connection and
// returns when the private key has been received. It consumes all expected messages.
func runFullKeyGen(t *testing.T, conn *websocket.Conn, flavor string) {
	t.Helper()
	steps := []struct {
		msg  models.InboundMessage
		want string
	}{
		{inbound(models.MsgSelectFlavor, flavor, ""), models.TypeParams},
		{inbound(models.MsgStepNext, "", models.StepGenerateRhoSigma), models.TypeRhoSigma},
		{inbound(models.MsgStepNext, "", models.StepGenerateMatrixA), models.TypeMatrixA},
		{inbound(models.MsgStepNext, "", models.StepGenerateVectors), models.TypeVectors},
		{inbound(models.MsgStepNext, "", models.StepComputeT), models.TypeTComputed},
		{inbound(models.MsgStepNext, "", models.StepSendPublicKey), models.TypePublicKeySent},
	}
	for _, s := range steps {
		send(t, conn, s.msg)
		msg := recv(t, conn)
		if msg.Type != s.want {
			t.Fatalf("keygen step %q: want %q got %q", s.msg.Step, s.want, msg.Type)
		}
	}
	// Consume the public_key_recv message.
	extra := recv(t, conn)
	if extra.Type != models.TypePublicKeyRecv {
		t.Fatalf("expected %q got %q", models.TypePublicKeyRecv, extra.Type)
	}
}

// TestSendMessageEncapsDecaps verifies the full encrypt → decrypt round-trip
// over the WebSocket for all three ML-KEM parameter sets.
func TestSendMessageEncapsDecaps(t *testing.T) {
	for _, flavor := range []string{models.Flavor512, models.Flavor768, models.Flavor1024} {
		t.Run(flavor, func(t *testing.T) {
			srv, conn := testServer(t)
			defer srv.Close()
			defer conn.Close()

			runFullKeyGen(t, conn, flavor)

			send(t, conn, models.InboundMessage{Type: models.MsgSendMessage})

			encMsg := recv(t, conn)
			if encMsg.Type != models.TypeEncryptResult {
				t.Fatalf("want %q got %q", models.TypeEncryptResult, encMsg.Type)
			}
			var ep models.EncryptResultPayload
			payloadAs(t, encMsg, &ep)
			if len(ep.Ciphertext) == 0 {
				t.Error("ciphertext is empty")
			}
			if len(ep.SharedSecret) != 64 { // 32 bytes → 64 hex chars
				t.Errorf("shared secret hex length: want 64 got %d", len(ep.SharedSecret))
			}

			decMsg := recv(t, conn)
			if decMsg.Type != models.TypeDecryptResult {
				t.Fatalf("want %q got %q", models.TypeDecryptResult, decMsg.Type)
			}
			var dp models.DecryptResultPayload
			payloadAs(t, decMsg, &dp)
			if !dp.Match {
				t.Errorf("shared secrets do not match: encaps=%s decaps=%s",
					ep.SharedSecret, dp.SharedSecret)
			}
			if ep.SharedSecret != dp.SharedSecret {
				t.Errorf("shared secret mismatch: encaps=%s decaps=%s",
					ep.SharedSecret, dp.SharedSecret)
			}
		})
	}
}

// TestSendMessageWithoutPublicKey verifies that send_message before key generation
// returns an error.
func TestSendMessageWithoutPublicKey(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	send(t, conn, inbound(models.MsgSelectFlavor, models.Flavor512, ""))
	recv(t, conn)

	send(t, conn, models.InboundMessage{Type: models.MsgSendMessage})
	msg := recv(t, conn)
	if msg.Type != models.TypeError {
		t.Fatalf("want %q got %q", models.TypeError, msg.Type)
	}
}

// TestSendMessageWithoutSession verifies send_message returns error when no session result.
func TestSendMessageWithoutSession(t *testing.T) {
	srv, conn := testServer(t)
	defer srv.Close()
	defer conn.Close()

	// No flavor selected at all.
	send(t, conn, models.InboundMessage{Type: models.MsgSendMessage})
	msg := recv(t, conn)
	if msg.Type != models.TypeError {
		t.Fatalf("want %q got %q", models.TypeError, msg.Type)
	}
}
