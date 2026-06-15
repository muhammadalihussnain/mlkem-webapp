package mlkem

import (
    "testing"
)

func TestNewParams(t *testing.T) {
    tests := []struct {
        name      string
        flavor    string
        wantK     int
        wantEta1  int
        wantEta2  int
        wantDu    int
        wantDv    int
        wantPkSize int
        wantSkSize int
        wantCtSize int
        wantErr   bool
    }{
        {"ML-KEM-512", "512", 2, 3, 2, 10, 4, 800, 768, 768, false},
        {"ML-KEM-768", "768", 3, 2, 2, 10, 4, 1184, 1152, 1088, false},
        {"ML-KEM-1024", "1024", 4, 2, 2, 11, 5, 1568, 1536, 1568, false},
        {"Unknown flavor", "unknown", 0, 0, 0, 0, 0, 0, 0, 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            params, err := NewParams(tt.flavor)
            if tt.wantErr {
                if err == nil {
                    t.Errorf("NewParams(%q) expected error, got nil", tt.flavor)
                }
                return
            }
            if err != nil {
                t.Errorf("NewParams(%q) unexpected error: %v", tt.flavor, err)
                return
            }
            
            // Test common constants
            if params.N != 256 {
                t.Errorf("N = %d, want 256", params.N)
            }
            if params.Q != 3329 {
                t.Errorf("Q = %d, want 3329", params.Q)
            }
            
            // Test flavor-specific values
            if params.K != tt.wantK {
                t.Errorf("K = %d, want %d", params.K, tt.wantK)
            }
            if params.Eta1 != tt.wantEta1 {
                t.Errorf("Eta1 = %d, want %d", params.Eta1, tt.wantEta1)
            }
            if params.Eta2 != tt.wantEta2 {
                t.Errorf("Eta2 = %d, want %d", params.Eta2, tt.wantEta2)
            }
            if params.Du != tt.wantDu {
                t.Errorf("Du = %d, want %d", params.Du, tt.wantDu)
            }
            if params.Dv != tt.wantDv {
                t.Errorf("Dv = %d, want %d", params.Dv, tt.wantDv)
            }
            if params.PkSize != tt.wantPkSize {
                t.Errorf("PkSize = %d, want %d", params.PkSize, tt.wantPkSize)
            }
            if params.SkSize != tt.wantSkSize {
                t.Errorf("SkSize = %d, want %d", params.SkSize, tt.wantSkSize)
            }
            if params.CtSize != tt.wantCtSize {
                t.Errorf("CtSize = %d, want %d", params.CtSize, tt.wantCtSize)
            }
        })
    }
}

func TestKeySizeConstants(t *testing.T) {
    // Verify sizes match FIPS 203 specifications
    flavors := []string{"512", "768", "1024"}
    expectedPkSizes := map[string]int{"512": 800, "768": 1184, "1024": 1568}
    expectedSkSizes := map[string]int{"512": 768, "768": 1152, "1024": 1536}
    expectedCtSizes := map[string]int{"512": 768, "768": 1088, "1024": 1568}
    
    for _, flavor := range flavors {
        params, err := NewParams(flavor)
        if err != nil {
            t.Fatalf("NewParams(%q) failed: %v", flavor, err)
        }
        
        if params.PkSize != expectedPkSizes[flavor] {
            t.Errorf("%s: PkSize = %d, want %d", flavor, params.PkSize, expectedPkSizes[flavor])
        }
        if params.SkSize != expectedSkSizes[flavor] {
            t.Errorf("%s: SkSize = %d, want %d", flavor, params.SkSize, expectedSkSizes[flavor])
        }
        if params.CtSize != expectedCtSizes[flavor] {
            t.Errorf("%s: CtSize = %d, want %d", flavor, params.CtSize, expectedCtSizes[flavor])
        }
    }
}
