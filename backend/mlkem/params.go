cat > backend/mlkem/params.go << 'EOF'
package mlkem

import "fmt"

type Params struct {
	N     int
	Q     int
	K     int
	Eta1  int
	Eta2  int
	Delta int

	Du     int
	Dv     int
	PkSize int
	SkSize int
	CtSize int
}

func NewParams(flavor string) (*Params, error) {
	switch flavor {

	case "512":
		return &Params{
			N:     256,
			Q:     3329,
			K:     2,
			Eta1:  3,
			Eta2:  2,
			Delta: 1 << 12,

			Du: 10,
			Dv: 4,

			PkSize: 800,
			SkSize: 768,  // Fixed: was 1632
			CtSize: 768,
		}, nil

	case "768":
		return &Params{
			N:     256,
			Q:     3329,
			K:     3,
			Eta1:  2,
			Eta2:  2,
			Delta: 1 << 12,

			Du: 10,
			Dv: 4,

			PkSize: 1184,
			SkSize: 1152,  // Fixed: was 2400
			CtSize: 1088,
		}, nil

	case "1024":
		return &Params{
			N:     256,
			Q:     3329,
			K:     4,
			Eta1:  2,
			Eta2:  2,
			Delta: 1 << 12,

			Du: 11,
			Dv: 5,

			PkSize: 1568,
			SkSize: 1536,  // Fixed: was 3168
			CtSize: 1568,
		}, nil

	default:
		return nil, fmt.Errorf("invalid ML-KEM flavor: %s", flavor)
	}
}
EOF
