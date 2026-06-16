package ws

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"ntt_verification/mlkem"
	"ntt_verification/models"
)

// ── WebSocket upgrade configuration ───────────────────────────────────────────

// sendBufSize is the number of outbound messages buffered per client before
// back-pressure causes the broadcaster to drop messages for that client.
const sendBufSize = 64

// writeWait is the maximum time allowed to write a single message to the peer.
const writeWait = 10 * time.Second

// pongWait is the maximum time to wait for a pong reply to a ping.
const pongWait = 60 * time.Second

// pingPeriod is how often the server sends a ping frame. Must be less than pongWait.
const pingPeriod = (pongWait * 9) / 10

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for local development. In production, restrict this.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ── Session state ──────────────────────────────────────────────────────────────

// session holds all ephemeral key-generation state for a single WebSocket connection.
type session struct {
	params *mlkem.Params
	result *mlkem.KeyGenResult
}

// ── Handler ────────────────────────────────────────────────────────────────────

// Handler upgrades an HTTP connection to WebSocket and runs the read/write pumps
// for a single client.
func Handler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws: upgrade error: %v", err)
			return
		}

		c := &client{conn: conn, send: make(chan []byte, sendBufSize)}
		hub.register(c)

		sess := &session{}

		go writePump(c)
		readPump(c, hub, sess)
	}
}

// readPump reads inbound messages from the client and drives the state machine.
// It runs in the calling goroutine; writePump runs in its own goroutine.
func readPump(c *client, hub *Hub, sess *session) {
	defer func() {
		hub.unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws: read error: %v", err)
			}
			return
		}

		var msg models.InboundMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			sendError(c, "invalid JSON: "+err.Error())
			continue
		}

		if err := handleMessage(c, hub, sess, msg); err != nil {
			sendError(c, err.Error())
		}
	}
}

// writePump drains the client's send channel and forwards messages to the WebSocket.
// It also sends periodic pings to detect dead connections.
func writePump(c *client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("ws: write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ── State machine ──────────────────────────────────────────────────────────────

// handleMessage dispatches an inbound message to the appropriate step handler.
func handleMessage(c *client, hub *Hub, sess *session, msg models.InboundMessage) error {
	switch msg.Type {
	case models.MsgSelectFlavor:
		return handleSelectFlavor(c, hub, sess, msg.Flavor)
	case models.MsgStepNext:
		return handleStepNext(c, hub, sess, msg.Step)
	case models.MsgSendMessage:
		return handleSendMessage(hub, sess)
	default:
		return nil // unknown types are silently ignored
	}
}

// handleSelectFlavor validates the flavor, builds Params, and broadcasts them.
func handleSelectFlavor(c *client, hub *Hub, sess *session, flavor string) error {
	params, err := mlkem.NewParams(flavor)
	if err != nil {
		return err
	}
	sess.params = params
	sess.result = nil // reset any previous run

	hub.Broadcast(mustMarshal(models.OutboundMessage{
		Type: models.TypeParams,
		Payload: models.ParamsPayload{
			Flavor: flavor,
			N:      params.N,
			Q:      params.Q,
			K:      params.K,
			Eta1:   params.Eta1,
			Eta2:   params.Eta2,
			Du:     params.Du,
			Dv:     params.Dv,
			PkSize: params.PkSize,
			SkSize: params.SkSize,
			CtSize: params.CtSize,
		},
	}))
	return nil
}

// handleStepNext advances the key-generation state machine by one step.
func handleStepNext(c *client, hub *Hub, sess *session, step string) error {
	if sess.params == nil {
		return errorf("no flavor selected; send select_flavor first")
	}

	switch step {
	case models.StepGenerateRhoSigma:
		return stepGenerateRhoSigma(hub, sess)
	case models.StepGenerateMatrixA:
		return stepGenerateMatrixA(hub, sess)
	case models.StepGenerateVectors:
		return stepGenerateVectors(hub, sess)
	case models.StepComputeT:
		return stepComputeT(hub, sess)
	case models.StepSendPublicKey:
		return stepSendPublicKey(hub, sess)
	default:
		return nil
	}
}

// stepGenerateRhoSigma generates a random seed, derives rho/sigma via G,
// and stores them in the session for subsequent steps.
func stepGenerateRhoSigma(hub *Hub, sess *session) error {
	seed := make([]byte, mlkem.SeedLen)
	if _, err := rand.Read(seed); err != nil {
		return errorf("failed to generate random seed: " + err.Error())
	}

	rho, sigma := mlkem.G(seed)

	// Initialise a partial result so subsequent steps can build on it.
	sess.result = &mlkem.KeyGenResult{
		Seed:  seed,
		Rho:   rho,
		Sigma: sigma,
	}

	hub.Broadcast(mustMarshal(models.OutboundMessage{
		Type: models.TypeRhoSigma,
		Payload: models.RhoSigmaPayload{
			Seed:  hex.EncodeToString(seed),
			Rho:   hex.EncodeToString(rho),
			Sigma: hex.EncodeToString(sigma),
		},
	}))
	return nil
}

// stepGenerateMatrixA samples the public matrix A from rho and broadcasts it.
func stepGenerateMatrixA(hub *Hub, sess *session) error {
	if sess.result == nil {
		return errorf("rho/sigma not yet generated; run generate_rho_sigma first")
	}

	A := mlkem.GenerateMatrixA(sess.result.Rho, sess.params.K)
	sess.result.A = A

	payload := models.MatrixAPayload{K: sess.params.K}
	for i := range payload.A {
		for j := range payload.A[i] {
			coeffs := make([]int32, mlkem.N)
			copy(coeffs, A[i][j][:])
			payload.A[i][j] = coeffs
		}
	}

	hub.Broadcast(mustMarshal(models.OutboundMessage{
		Type:    models.TypeMatrixA,
		Payload: payload,
	}))
	return nil
}

// stepGenerateVectors samples s and e, broadcasts the raw byte streams then the vectors.
func stepGenerateVectors(hub *Hub, sess *session) error {
	if sess.result == nil {
		return errorf("rho/sigma not yet generated; run generate_rho_sigma first")
	}

	s, e := mlkem.GenerateSecretAndError(
		sess.result.Sigma,
		sess.params.K,
		sess.params.Eta1,
		sess.params.Eta2,
	)
	sess.result.S = s
	sess.result.E = e

	// Broadcast vector payload.
	vp := models.VectorsPayload{K: sess.params.K}
	for i := 0; i < sess.params.K; i++ {
		sCoeffs := make([]int32, mlkem.N)
		eCoeffs := make([]int32, mlkem.N)
		copy(sCoeffs, s[i][:])
		copy(eCoeffs, e[i][:])
		vp.S[i] = sCoeffs
		vp.E[i] = eCoeffs
	}

	hub.Broadcast(mustMarshal(models.OutboundMessage{
		Type:    models.TypeVectors,
		Payload: vp,
	}))
	return nil
}

// stepComputeT computes t = NTT^{-1}(A·NTT(s)) + e and broadcasts the result.
func stepComputeT(hub *Hub, sess *session) error {
	if sess.result == nil {
		return errorf("vectors not yet generated; run generate_vectors first")
	}

	// Use KeyGen internals: transform s, multiply, inverse-transform, add e.
	var sHat [mlkem.MaxMatrixDim][mlkem.N]int32
	for i := 0; i < sess.params.K; i++ {
		sHat[i] = mlkem.NTTForward(sess.result.S[i])
	}

	tHat := mlkem.MatVecMulNTT(sess.result.A, sHat, sess.params.K)
	var t [mlkem.MaxMatrixDim][mlkem.N]int32
	for i := 0; i < sess.params.K; i++ {
		tPoly := mlkem.NTTInverse(tHat[i])
		t[i] = mlkem.PolyAdd(tPoly, sess.result.E[i])
	}
	sess.result.T = t

	tp := models.TComputedPayload{K: sess.params.K}
	for i := 0; i < sess.params.K; i++ {
		coeffs := make([]int32, mlkem.N)
		copy(coeffs, t[i][:])
		tp.T[i] = coeffs
	}

	hub.Broadcast(mustMarshal(models.OutboundMessage{
		Type:    models.TypeTComputed,
		Payload: tp,
	}))
	return nil
}

// stepSendPublicKey encodes pk and sk, then broadcasts both.
func stepSendPublicKey(hub *Hub, sess *session) error {
	if sess.result == nil {
		return errorf("t not yet computed; run compute_t first")
	}

	pk := mlkem.EncodePolyVec(sess.result.T, sess.params.K)
	pk = append(pk, sess.result.Rho...)
	sk := mlkem.EncodePolyVec(sess.result.S, sess.params.K)

	sess.result.PublicKey = pk
	sess.result.PrivateKey = sk

	hub.Broadcast(mustMarshal(models.OutboundMessage{
		Type: models.TypePublicKeySent,
		Payload: models.PublicKeyPayload{
			PublicKey:     hex.EncodeToString(pk),
			PublicKeySize: len(pk),
		},
	}))
	hub.Broadcast(mustMarshal(models.OutboundMessage{
		Type: models.TypePublicKeyRecv,
		Payload: models.PrivateKeyPayload{
			PrivateKey:     hex.EncodeToString(sk),
			PrivateKeySize: len(sk),
		},
	}))
	return nil
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// handleSendMessage runs Encapsulate on the current public key, then immediately
// runs Decapsulate on the server side and broadcasts both results.
// This demonstrates the full encrypt → decrypt round-trip in one step.
func handleSendMessage(hub *Hub, sess *session) error {
	if sess.result == nil || len(sess.result.PublicKey) == 0 {
		return errorf("public key not yet available; complete key generation first")
	}

	enc, err := mlkem.Encapsulate(sess.result.PublicKey, sess.params)
	if err != nil {
		return errorf("encapsulate failed: " + err.Error())
	}

	hub.Broadcast(mustMarshal(models.OutboundMessage{
		Type: models.TypeEncryptResult,
		Payload: models.EncryptResultPayload{
			Ciphertext:     hex.EncodeToString(enc.Ciphertext),
			CiphertextSize: len(enc.Ciphertext),
			SharedSecret:   hex.EncodeToString(enc.SharedSecret),
			Message:        hex.EncodeToString(enc.Message),
		},
	}))

	ss, err := mlkem.Decapsulate(sess.result.PrivateKey, enc.Ciphertext, sess.result.PublicKey, sess.params)
	if err != nil {
		return errorf("decapsulate failed: " + err.Error())
	}

	hub.Broadcast(mustMarshal(models.OutboundMessage{
		Type: models.TypeDecryptResult,
		Payload: models.DecryptResultPayload{
			SharedSecret: hex.EncodeToString(ss),
			Match:        mlkem.ConstantTimeEqual(enc.SharedSecret, ss),
		},
	}))
	return nil
}

// sendError sends a TypeError message directly to a single client (not broadcast).
func sendError(c *client, msg string) {
	data := mustMarshal(models.OutboundMessage{
		Type:    models.TypeError,
		Payload: models.ErrorPayload{Message: msg},
	})
	select {
	case c.send <- data:
	default:
	}
}

// mustMarshal serialises v to JSON, panicking on error (only possible with
// non-serialisable types, which we control).
func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic("ws: failed to marshal message: " + err.Error())
	}
	return b
}

// errorf returns a plain error with a formatted message.
func errorf(msg string) error {
	return &wsError{msg: msg}
}

type wsError struct{ msg string }

func (e *wsError) Error() string { return e.msg }
