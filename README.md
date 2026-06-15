# ML-KEM Web Application

[![CI](https://github.com/muhammadalihussnain/mlkem-webapp/actions/workflows/ci.yml/badge.svg)](https://github.com/muhammadalihussnain/mlkem-webapp/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/muhammadalihussnain/mlkem-webapp/graph/badge.svg)](https://codecov.io/gh/muhammadalihussnain/mlkem-webapp)

## ML-KEM (CRYSTALS-Kyber) Post-Quantum Cryptography Web Application

This application implements ML-KEM (formerly CRYSTALS-Kyber) as specified in FIPS 203.

## Project Structure
.
├── backend/ # Go backend with ML-KEM implementation
├── frontend/ # React TypeScript frontend
├── .github/ # GitHub Actions CI/CD
└── Makefile # Build and test automation


## Current Status (Day 1)

✅ Repository setup  
✅ CI pipeline configured  
✅ ML-KEM parameters defined (FIPS 203 compliant)  
✅ 100% test coverage for parameters module  

## Running Locally

```bash
# Run all tests
make test

# Check coverage
make coverage


### Step 11: Run Verification Gates
```bash
# Run backend tests
cd backend
go test ./mlkem/... -v -coverprofile=coverage.out
go tool cover -func=coverage.out  # Should show 100% for params.go

# Run from root
cd ..
make test
