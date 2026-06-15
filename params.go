package mlkem

import "fmt"

type Params struct {
    N      int // 256 for all ML-KEM variants
    Q      int // 3329 for all ML-KEM variants
    K      int // Security level parameter (2,3,4 for 512,768,1024)
    Eta1   int // Noise parameter for private key
    Eta2   int // Noise parameter for ciphertext
    Du     int // Compression parameter for ciphertext u
    Dv     int // Compression parameter for ciphertext v
    PkSize int
    SkSize int
    CtSize int
}

func NewParams(flavor string) (*Params, error) {
    switch flavor {
    case "512":
        return &Params{
            N:      256,
            Q:      3329,
            K:      2,
            Eta1:   3,
            Eta2:   2,
            Du:     10,
            Dv:     4,
            PkSize: 800,
            SkSize: 768,
            CtSize: 768,
        }, nil
    case "768":
        return &Params{
            N:      256,
            Q:      3329,
            K:      3,
            Eta1:   2,
            Eta2:   2,
            Du:     10,
            Dv:     4,
            PkSize: 1184,
            SkSize: 1152,
            CtSize: 1088,
        }, nil
    case "1024":
        return &Params{
            N:      256,
            Q:      3329,
            K:      4,
            Eta1:   2,
            Eta2:   2,
            Du:     11,
            Dv:     5,
            PkSize: 1568,
            SkSize: 1536,
            CtSize: 1568,
        }, nil
    default:
        return nil, fmt.Errorf("unknown flavor: %s, must be 512, 768, or 1024", flavor)
    }
}
