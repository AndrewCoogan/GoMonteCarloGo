package core

import (
	"math"
	"math/rand/v2"
	"testing"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

const (
	mu_a           = 0.08
	mu_b           = 0.10
	mu_c           = 0.12
	sigma_a        = 0.15
	sigma_b        = 0.20
	sigma_c        = 0.25
	corr_ab        = 0.5
	corr_ac        = 0.0
	corr_bc        = 0.0
	days_in_a_year = 252
)

// TestSupportingGenerators ensures the math is correct for supporting testing functionality
func TestSupportingGenerators(t *testing.T) {
	nSamples := days_in_a_year * 500
	returns := generateMockReturns(t, nSamples)

	// verify return correlation
	corr01 := stat.Correlation(returns[0], returns[1], nil)
	corr02 := stat.Correlation(returns[0], returns[2], nil)
	corr12 := stat.Correlation(returns[1], returns[2], nil)

	correlationTolerance := 0.01
	if math.Abs(corr01-corr_ab) > correlationTolerance {
		t.Errorf("Corr(Asset0, Asset1): expected %.4f, got %.4f", corr_ab, corr01)
	}

	if math.Abs(corr02-corr_ac) > correlationTolerance {
		t.Errorf("Corr(Asset0, Asset2): expected %.4f, got %.4f", corr_ac, corr02)
	}

	if math.Abs(corr12-corr_bc) > correlationTolerance {
		t.Errorf("Corr(Asset1, Asset2): expected %.4f, got %.4f", corr_bc, corr12)
	}

	// verify mean
	mu1 := stat.Mean(returns[0], nil) * days_in_a_year
	mu2 := stat.Mean(returns[1], nil) * days_in_a_year
	mu3 := stat.Mean(returns[2], nil) * days_in_a_year
	drift_adjusted_mu1 := mu_a - 0.5*math.Pow(sigma_a, 2)
	drift_adjusted_mu2 := mu_b - 0.5*math.Pow(sigma_b, 2)
	drift_adjusted_mu3 := mu_c - 0.5*math.Pow(sigma_c, 2)

	muTolerance := 0.01
	if math.Abs(mu1-drift_adjusted_mu1) > muTolerance {
		t.Errorf("Mu(Asset0): expected %.4f, got %.4f", drift_adjusted_mu1, mu1)
	}

	if math.Abs(mu2-drift_adjusted_mu2) > muTolerance {
		t.Errorf("Mu(Asset1): expected %.4f, got %.4f", drift_adjusted_mu2, mu2)
	}

	if math.Abs(mu3-drift_adjusted_mu3) > muTolerance {
		t.Errorf("Mu(Asset2): expected %.4f, got %.4f", drift_adjusted_mu3, mu3)
	}

	// verify standard deviation
	sigma1 := stat.StdDev(returns[0], nil) * math.Sqrt(days_in_a_year)
	sigma2 := stat.StdDev(returns[1], nil) * math.Sqrt(days_in_a_year)
	sigma3 := stat.StdDev(returns[2], nil) * math.Sqrt(days_in_a_year)

	sigmaTolerance := 0.01
	if math.Abs(sigma1-sigma_a) > sigmaTolerance {
		t.Errorf("Sigma(Asset0): expected %.4f, got %.4f", sigma_a, sigma1)
	}

	if math.Abs(sigma2-sigma_b) > sigmaTolerance {
		t.Errorf("Sigma(Asset1): expected %.4f, got %.4f", sigma_b, sigma2)
	}

	if math.Abs(sigma3-sigma_c) > sigmaTolerance {
		t.Errorf("Sigma(Asset2): expected %.4f, got %.4f", sigma_c, sigma3)
	}

	prices := generateMockStockPrices(t, returns) // seeds are same, should the same underlying data
	for asset := range returns {
		for day := range len(returns[asset]) {
			calculatedReturn := math.Log(prices[asset][day+1] / prices[asset][day])
			diff := math.Abs(returns[asset][day] - calculatedReturn)
			if diff > 1e-10 {
				t.Errorf("Asset %d, Day %d: return mismatch (diff: %.2e)", asset, day, diff)
			}
		}
	}
}

// TestBasicSetup verifies that SharedMatrices are created correctly
func TestBasicSetup(t *testing.T) {
	nSamples := days_in_a_year * 100
	returns := generateMockReturns(t, nSamples)
 
	shared, err := GetStatisticalResources(returns, StandardNormal, 0)
	if err != nil {
		t.Fatalf("Failed to create SharedMatrices: %v", err)
	}
 
	if len(shared.MeanReturns) != 3 {
		t.Errorf("Expected 3 assets, got %d", len(shared.MeanReturns))
	}
 
	expectedMeans := []float64{mu_a, mu_b, mu_c}
	for i, expected := range expectedMeans {
		if math.Abs(shared.MeanReturns[i]-expected) > 0.02 {
			t.Errorf("Asset %d: expected mean ~%.2f, got %.4f", i, expected, shared.MeanReturns[i])
		}
	}
 
	expectedStds := []float64{sigma_a, sigma_b, sigma_c}
	for i, expected := range expectedStds {
		if math.Abs(shared.StandardDeviation[i]-expected) > 0.02 {
			t.Errorf("Asset %d: expected std ~%.2f, got %.4f", i, expected, shared.StandardDeviation[i])
		}
	}
 }
 
 

// Helper: Generate one year of mock daily historical returns
func generateMockReturns(t *testing.T, n int) [][]float64 {
	t.Helper()

	nAssets := 3
	corrData := []float64{
		1.0, corr_ab, corr_ac,
		corr_ab, 1.0, corr_bc,
		corr_ac, corr_bc, 1.0,
	}

	corrMatrix := mat.NewSymDense(nAssets, corrData)
	var chol mat.Cholesky
	if ok := chol.Factorize(corrMatrix); !ok {
		t.Fatalf("Correlation matrix is not positive definite")
	}

	L := new(mat.TriDense)
	chol.LTo(L)

	src := rand.NewPCG(42, 0)
	normalDist := distuv.Normal{Mu: 0, Sigma: 1, Src: src}

	asset_a := make([]float64, n)
	asset_b := make([]float64, n)
	asset_c := make([]float64, n)

	z := make([]float64, nAssets)
	for sim := range n {
		for i := range nAssets {
			z[i] = normalDist.Rand()
		}

		zVec := mat.NewVecDense(nAssets, z)
		correlatedZ := mat.NewVecDense(nAssets, nil)
		correlatedZ.MulVec(L, zVec)

		asset_a[sim] = calculateLogNormalReturn(t, mu_a, sigma_a, correlatedZ.AtVec(0), days_in_a_year)
		asset_b[sim] = calculateLogNormalReturn(t, mu_b, sigma_b, correlatedZ.AtVec(1), days_in_a_year)
		asset_c[sim] = calculateLogNormalReturn(t, mu_c, sigma_c, correlatedZ.AtVec(2), days_in_a_year)
	}

	return [][]float64{asset_a, asset_b, asset_c}
}

// Helper: Centralized way to calculate log normal returns
func calculateLogNormalReturn(t *testing.T, mu, sigma, rng, normalization float64) float64 {
	t.Helper()
	return (mu-0.5*math.Pow(sigma, 2))/normalization + (sigma * rng / math.Sqrt(normalization))
}

// Helper: Generate one year of mock daily historical prices
func generateMockStockPrices(t *testing.T, returns [][]float64) [][]float64 {
	t.Helper()

	nSamples := len(returns[0])
	asset_a := make([]float64, nSamples+1)
	asset_b := make([]float64, nSamples+1)
	asset_c := make([]float64, nSamples+1)

	asset_a[0] = 100 // initial value for asset a
	asset_b[0] = 50  // asset b
	asset_c[0] = 200 // and asset c

	for sim := range nSamples {
		asset_a[sim+1] = asset_a[sim] * math.Exp(returns[0][sim])
		asset_b[sim+1] = asset_b[sim] * math.Exp(returns[1][sim])
		asset_c[sim+1] = asset_c[sim] * math.Exp(returns[2][sim])
	}

	return [][]float64{asset_a, asset_b, asset_c}
}
