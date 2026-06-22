import { useWebSocket } from './hooks/useWebSocket';
import { IntroSection } from './components/IntroSection';
import { FlowDiagram } from './components/FlowDiagram';
import { FlavorSelector } from './components/FlavorSelector';
import { ParameterPanel } from './components/ParameterPanel';
import { StepButton } from './components/StepButton';
import { RhoSigmaPanel } from './components/RhoSigmaPanel';
import { MatrixDisplay } from './components/MatrixDisplay';
import { VectorDisplay } from './components/VectorDisplay';
import { KeyPanel } from './components/KeyPanel';
import { MessagePanel } from './components/MessagePanel';
import { EventLog } from './components/EventLog';
import { STEPS } from './types/protocol';
import './App.css';

const STEP_TOOLTIPS: Record<string, string> = {
  generate_rho_sigma:
    'Samples 32 bytes of OS/HW entropy (seed), then hashes it with G to produce ' +
    '\u03c1 (public nonce \u2192 expands into matrix A) and \u03c3 (secret nonce \u2192 samples s and e).',
  generate_matrix_A:
    'Expands \u03c1 into a k\u00d7k matrix A of polynomials in the NTT domain using ' +
    'SHAKE-128 (XOF). A is fully determined by \u03c1 and is part of the public key.',
  generate_vectors:
    'Samples secret vector s and error vector e from the centered binomial ' +
    'distribution B\u03b7\u2081 using \u03c3 as a PRF seed. Both are kept private.',
  compute_t:
    'Computes the public vector t = A\u00b7s + e in the NTT domain (mod q). ' +
    'This combines the public matrix A with the secret s and small noise e.',
  send_public_key:
    'Encodes (\u03c1, t) into the standard ML-KEM public key byte string ' +
    'of size |pk| bytes, ready to be shared with the encapsulator.',
};

function App() {
  const { isConnected, state, sendMessage, resetState } = useWebSocket();

  // Lock to light mode always
  document.documentElement.setAttribute('data-theme', 'light');

  const handleFlavorSelect = (flavor: string) => {
    sendMessage({ type: 'select_flavor', flavor });
  };

  const handleStepNext = (stepId: string) => {
    sendMessage({ type: 'step_next', step: stepId });
  };

  const handleSendMessage = () => {
    sendMessage({ type: 'send_message' });
  };

  const handleReset = () => {
    sendMessage({ type: 'reset' });
    resetState();
  };

  return (
    <div className="app">
      <header className="app__header">
        <h1>
          ML-KEM Post-Quantum Key Encapsulation&nbsp;
          <a href="https://doi.org/10.6028/NIST.FIPS.203" target="_blank"
            rel="noopener noreferrer" className="app__header-fips">FIPS&nbsp;203</a>
        </h1>
        <div className="app__status">
          <span aria-label={`connection status: ${isConnected ? 'connected' : 'disconnected'}`}>
            {isConnected ? '\uD83D\uDFE2 Connected' : '\uD83D\uDD34 Disconnected'}
          </span>
          <button type="button" onClick={handleReset} className="app__reset">
            Reset
          </button>
        </div>
      </header>

      <IntroSection />
      <FlowDiagram />

      <div className="demo-divider">
        <div className="demo-divider__line" />
        <div className="demo-divider__label">Interactive Demo</div>
        <div className="demo-divider__line" />
      </div>

      <section className="demo-instructions" aria-labelledby="demo-instructions-heading">
        <h2 id="demo-instructions-heading" className="demo-instructions__heading">How to use this demo</h2>
        <ol className="demo-instructions__list">
          <li>
            <strong>Pick a security level</strong> — choose ML-KEM-512, 768, or 1024. This determines the key sizes and security margin.
          </li>
          <li>
            <strong>Generate ρ and σ</strong> — click <em>Step 1</em> to sample 32 bytes of entropy and hash them into two seeds: ρ (used to build the public matrix A) and σ (used to sample the secret vectors).
          </li>
          <li>
            <strong>Expand matrix A</strong> — click <em>Step 2</em> to expand ρ into a k×k matrix of polynomials via SHAKE-128. This matrix is part of the public key.
          </li>
          <li>
            <strong>Sample secret vectors s and e</strong> — click <em>Step 3</em> to draw the private secret vector <em>s</em> and small error vector <em>e</em> from a centered binomial distribution seeded by σ.
          </li>
          <li>
            <strong>Compute public vector t</strong> — click <em>Step 4</em> to compute <em>t = A·s + e</em> in the NTT domain. This combines the public matrix with the private secret and noise.
          </li>
          <li>
            <strong>Build the public key</strong> — click <em>Step 5</em> to encode (ρ, t) into the final encapsulation key (ek) byte string. Your private key (dk) is derived and stored server-side.
          </li>
          <li>
            <strong>Run key exchange</strong> — once the public key is ready, use the <em>Key Exchange</em> section to encapsulate and decapsulate a shared secret.
          </li>
        </ol>
      </section>

      {/* ── Security level + parameters ──────────────────────────────────── */}
      <div className="demo-setup">
        <section className="card demo-setup__flavor" aria-labelledby="section-flavor">
          <h2 id="section-flavor">1 · Security Level</h2>
          <FlavorSelector
            onSelect={handleFlavorSelect}
            currentFlavor={state.flavor}
            disabled={!isConnected}
          />
        </section>
        <section className="card demo-setup__params" aria-labelledby="section-params">
          <h2 id="section-params">2 · Parameters</h2>
          <ParameterPanel params={state.params} />
        </section>
      </div>

      {/* ── Key generation steps — result appears directly below each button ── */}
      <section className="card" aria-labelledby="section-steps">
        <h2 id="section-steps">3 · Key Generation Steps</h2>

        {/* Step 1 — ρ / σ */}
        <div className="step-block">
          <StepButton
            label={STEPS[0].label}
            onClick={() => handleStepNext(STEPS[0].id)}
            disabled={!isConnected || !state.flavor}
            stepNumber={1}
            currentStep={state.currentStep}
            tooltip={STEP_TOOLTIPS[STEPS[0].id] ?? ''}
          />
          {state.rhoSigma && (
            <div className="step-result" aria-label="ρ σ result">
              <div className="step-result__label">Seed · ρ · σ</div>
              <RhoSigmaPanel rhoSigma={state.rhoSigma} />
            </div>
          )}
        </div>

        {/* Step 2 — Matrix A */}
        <div className="step-block">
          <StepButton
            label={STEPS[1].label}
            onClick={() => handleStepNext(STEPS[1].id)}
            disabled={!isConnected || !state.flavor}
            stepNumber={2}
            currentStep={state.currentStep}
            tooltip={STEP_TOOLTIPS[STEPS[1].id] ?? ''}
          />
          {state.matrixA && (
            <div className="step-result" aria-label="matrix A result">
              <div className="step-result__label">Matrix Â ({state.matrixA.k}×{state.matrixA.k})</div>
              <MatrixDisplay matrix={state.matrixA} />
            </div>
          )}
        </div>

        {/* Step 3 — s / e vectors */}
        <div className="step-block">
          <StepButton
            label={STEPS[2].label}
            onClick={() => handleStepNext(STEPS[2].id)}
            disabled={!isConnected || !state.flavor}
            stepNumber={3}
            currentStep={state.currentStep}
            tooltip={STEP_TOOLTIPS[STEPS[2].id] ?? ''}
          />
          {state.vectors && (
            <div className="step-result" aria-label="s e vectors result">
              <div className="step-result__label">Vectors s / e</div>
              <VectorDisplay vectors={state.vectors} tComputed={null} />
            </div>
          )}
        </div>

        {/* Step 4 — t vector */}
        <div className="step-block">
          <StepButton
            label={STEPS[3].label}
            onClick={() => handleStepNext(STEPS[3].id)}
            disabled={!isConnected || !state.flavor}
            stepNumber={4}
            currentStep={state.currentStep}
            tooltip={STEP_TOOLTIPS[STEPS[3].id] ?? ''}
          />
          {state.tComputed && (
            <div className="step-result" aria-label="t vector result">
              <div className="step-result__label">Public vector t</div>
              <VectorDisplay vectors={state.vectors} tComputed={state.tComputed} />
            </div>
          )}
        </div>

        {/* Step 5 — Public key */}
        <div className="step-block">
          <StepButton
            label={STEPS[4].label}
            onClick={() => handleStepNext(STEPS[4].id)}
            disabled={!isConnected || !state.flavor}
            stepNumber={5}
            currentStep={state.currentStep}
            tooltip={STEP_TOOLTIPS[STEPS[4].id] ?? ''}
          />
          {state.publicKey && (
            <div className="step-result" aria-label="public key result">
              <div className="step-result__label">Encapsulation Key (ek)</div>
              <KeyPanel publicKey={state.publicKey} />
            </div>
          )}
        </div>

      </section>

      {/* ── Key exchange ─────────────────────────────────────────────────── */}
      <section className="card" aria-labelledby="section-msg">
        <h2 id="section-msg">4 · Key Exchange</h2>
        <MessagePanel
          onSendMessage={handleSendMessage}
          encryptResult={state.encryptResult}
          decryptResult={state.decryptResult}
          disabled={!isConnected || !state.publicKey}
        />
      </section>

      {/* ── Event log (collapsed, secondary) ─────────────────────────────── */}
      <section className="card card--log" aria-labelledby="section-log">
        <h2 id="section-log">Event Log</h2>
        <EventLog events={state.events} />
      </section>

    </div>
  );
}

export default App;

