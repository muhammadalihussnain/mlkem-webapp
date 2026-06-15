import { useState } from 'react'
import './App.css'

function App() {
  const [count, setCount] = useState(0)

  return (
    <>
      <h1>ML-KEM Post-Quantum Cryptography</h1>
      <div className="card">
        <button onClick={() => setCount((count) => count + 1)}>
          count is {count}
        </button>
        <p>
          ML-KEM (CRYSTALS-Kyber) Web Application - FIPS 203 Compliant
        </p>
      </div>
    </>
  )
}

export default App
