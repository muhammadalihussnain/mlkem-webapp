import React from 'react';
import { useWebSocket } from './hooks/useWebSocket';
import { FlavorSelector } from './components/FlavorSelector';
import { ParameterPanel } from './components/ParameterPanel';
import { StepButton } from './components/StepButton';
import { MatrixDisplay } from './components/MatrixDisplay';
import { VectorDisplay } from './components/VectorDisplay';
import { KeyPanel } from './components/KeyPanel';
import { MessagePanel } from './components/MessagePanel';
import { EventLog } from './components/EventLog';
import { STEPS } from './types/protocol';
import './App.css';

function App() {
  const { isConnected, state, sendMessage, resetState } = useWebSocket();

  const handleFlavorSelect = (flavor: string) => {
    sendMessage({ type: 'select_flavor', flavor });
  };

  const handleStepNext = (stepId: string) => {
    sendMessage({ type: 'step_next', step: stepId });
  };

  const handleSendMessage = () => {
    sendMessage({ type: 'send_message' });
  };

  return (
    <div className="app">
      <header className="app__header">
        <h1>ML-KEM — Post-Quantum Key Encapsulation</h1>
        <div className="app__status">
          <span aria-label={`connection status: ${isConnected ? 'connected' : 'disconnected'}`}>
            {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
          </span>
          <button type="button" onClick={resetState} className="app__reset">
            Reset
          </button>
        </div>
      </header>

      <div className="app__layout">
        {/* ── Left column ── */}
        <div className="app__left">

          <section className="card" aria-labelledby="section-flavor">
            <h2 id="section-flavor">1 · Security Level</h2>
            <FlavorSelector
              onSelect={handleFlavorSelect}
              currentFlavor={state.flavor}
              disabled={!isConnected}
            />
          </section>

          <section className="card" aria-labelledby="section-params">
            <h2 id="section-params">2 · Parameters</h2>
            <ParameterPanel params={state.params} />
          </section>

          <section className="card" aria-labelledby="section-steps">
            <h2 id="section-steps">3 · Key Generation Steps</h2>
            {STEPS.map((step, idx) => (
              <StepButton
                key={step.id}
                label={step.label}
                onClick={() => handleStepNext(step.id)}
                disabled={!isConnected || !state.flavor}
                stepNumber={idx + 1}
                currentStep={state.currentStep}
              />
            ))}
          </section>

          <section className="card" aria-labelledby="section-key">
            <h2 id="section-key">4 · Public Key</h2>
            <KeyPanel publicKey={state.publicKey} />
          </section>

          <section className="card" aria-labelledby="section-msg">
            <h2 id="section-msg">5 · Key Exchange</h2>
            <MessagePanel
              onSendMessage={handleSendMessage}
              encryptResult={state.encryptResult}
              decryptResult={state.decryptResult}
              disabled={!isConnected || !state.publicKey}
            />
          </section>

        </div>

        {/* ── Right column ── */}
        <div className="app__right">

          <section className="card" aria-labelledby="section-matrix">
            <h2 id="section-matrix">Matrix A</h2>
            <MatrixDisplay matrix={state.matrixA} />
          </section>

          <section className="card" aria-labelledby="section-vectors">
            <h2 id="section-vectors">Vectors s / e / t</h2>
            <VectorDisplay vectors={state.vectors} tComputed={state.tComputed} />
          </section>

          <section className="card card--log" aria-labelledby="section-log">
            <h2 id="section-log">Event Log</h2>
            <EventLog events={state.events} />
          </section>

        </div>
      </div>
    </div>
  );
}

export default App;
