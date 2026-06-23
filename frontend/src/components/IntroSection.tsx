const FIPS_URL = 'https://doi.org/10.6028/NIST.FIPS.203';

function Fips203Link() {
  return (
    <a href={FIPS_URL} target="_blank" rel="noopener noreferrer" className="fips-link">
      FIPS&nbsp;203
    </a>
  );
}

const ANALOGIES = [
  {
    icon: '🔐',
    title: 'What is ML-KEM?',
    body: (
      <>
        ML-KEM is a post-quantum cryptographic algorithm standardized by NIST (<Fips203Link />).
        It uses structured lattices and the Learning with Errors (LWE) problem to securely
        exchange symmetric keys, designed to resist attacks from both classical and quantum computers.
      </>
    ),
  },
  {
    icon: '📦',
    title: 'The locked box analogy',
    body: (
      <>
        Alice makes a special locked box and sends it to Bob (the public key). Bob puts a secret
        note inside, locks it, and sends it back. Only Alice has the key to open it. Now they
        both know the secret and nobody else does.
      </>
    ),
  },
  {
    icon: '⚡',
    title: 'Why do we need it now?',
    body: (
      <>
        Today&apos;s encryption can be broken by future quantum computers. ML-KEM is built on
        hard math problems that even quantum computers can&apos;t solve quickly, so your data
        stays safe for decades to come.
      </>
    ),
  },
  {
    icon: '✅',
    title: 'Who approved it?',
    body: (
      <>
        NIST standardised ML-KEM in 2024 as <Fips203Link />. It is already being adopted by
        major browsers, operating systems, and cloud providers worldwide.
      </>
    ),
  },
];

const VARIANTS = [
  {
    name: 'ML-KEM-512',
    level: 'Level 1-equivalent to AES-128',
    desc: 'Smallest key and ciphertext sizes. Suitable for most applications and constrained environments.',
    ek: '800 bytes', dk: '1632 bytes',
  },
  {
    name: 'ML-KEM-768',
    level: 'Level 3-equivalent to AES-192',
    desc: 'Balanced security and performance. The recommended default for general-purpose use by NIST.',
    ek: '1184 bytes', dk: '2400 bytes',
  },
  {
    name: 'ML-KEM-1024',
    level: 'Level 5-equivalent to AES-256',
    desc: 'Highest security margin. Largest keys. For long-term sensitive data and high-assurance systems.',
    ek: '1568 bytes', dk: '3168 bytes',
  },
];

function StyledSubtitle() {
  const words = ['Module-', 'Lattice', 'Key', 'Encapsulation', 'Mechanism'];
  return (
    <span className="intro__title-sub">
      {words.map((word, i) => (
        <span key={i}>
          <strong className="intro__title-initial">{word[0]}</strong>
          {word.slice(1)}
          {i < words.length - 1 ? ' ' : ''}
        </span>
      ))}
    </span>
  );
}

export function IntroSection() {
  return (
    <section className="intro" aria-labelledby="intro-heading">
      <div className="intro__hero">
        <div className="intro__badge">
          <Fips203Link /> · Post-Quantum Cryptography
        </div>
        <h1 id="intro-heading" className="intro__title">
          ML-KEM
          <StyledSubtitle />
        </h1>
        <p className="intro__lead">
          The next-generation standard for sharing secrets securely, even against quantum computers.
          Explore how it works, step by step, right in your browser.
        </p>
      </div>

      {/* What / why / who cards */}
      <div className="intro__cards">
        {ANALOGIES.map(({ icon, title, body }) => (
          <div key={title} className="intro__card">
            <div className="intro__card-icon-wrap">
              <span className="intro__card-icon">{icon}</span>
            </div>
            <h3 className="intro__card-title">{title}</h3>
            <p className="intro__card-body">{body}</p>
          </div>
        ))}
      </div>

      {/* Three security levels + variants table combined */}
      <div className="intro__variants">
        <h2 className="intro__variants-heading">Three Security Levels</h2>
        <p className="intro__variants-lead">
          ML-KEM comes in three variants standardised in <Fips203Link />, each offering a
          different trade-off between key size and security margin.
        </p>
        <div className="intro__variants-grid">
          {VARIANTS.map(v => (
            <div key={v.name} className="intro__variant-card">
              <div className="intro__variant-name">{v.name}</div>
              <div className="intro__variant-level">{v.level}</div>
              <p className="intro__variant-desc">{v.desc}</p>
              <div className="intro__variant-sizes">
                <span><strong>Encryption Key:</strong> {v.ek}</span>
                <span><strong>Decryption Key:</strong> {v.dk}</span>
              </div>
            </div>
          ))}
        </div>

        {/* Security variants comparison table */}
        <div className="intro__variants-table-wrap" aria-label="ML-KEM security variants comparison">
          <table className="intro__variants-table">
            <thead>
              <tr>
                <th>Variant</th>
                <th>Security Level</th>
                <th>Encryption Key (ek)</th>
                <th>Decryption Key (dk)</th>
              </tr>
            </thead>
            <tbody>
              {VARIANTS.map(v => (
                <tr key={v.name}>
                  <td className="intro__variants-table__name">{v.name}</td>
                  <td>{v.level.replace(' \u2014 ', ' \u00b7 ')}</td>
                  <td>{v.ek}</td>
                  <td>{v.dk}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}
