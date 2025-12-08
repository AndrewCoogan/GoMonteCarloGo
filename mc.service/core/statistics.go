package core

import (
	"fmt"
	"math/rand/v2"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"

	ex "mc.data/extensions"
)

const (
	StandardNormal = iota
	StudentT
)

type StatisticalResources struct {
	CovMatrix         *mat.SymDense // covariance matrix for std normal dist
	CorrMatrix        *mat.SymDense // correlation matrix for student t dist
	CholeskyL         *mat.TriDense // cholesky of covariance (std normal dist)
	CholeskyCorrL     *mat.TriDense // cholesky of correlation (student t dist)
	MeanReturns       []float64
	StandardDeviation []float64
	DistType          int
	Df                float64
}

// Used for parallelization, will have shared materials to minimize memory usage
type WorkerResources struct {
	*StatisticalResources           // embed read only shared data
	rng                   *rand.PCG // worker-specific RNG
}

// Called in the go routine and have seeds respectively set for each
func NewWorkerResources(shared *StatisticalResources, seed uint64) *WorkerResources {
	rng := new(rand.PCG)
	if seed != 0 {
		rng = rand.NewPCG(seed, 0)
	}

	return &WorkerResources{
		StatisticalResources: shared,
		rng:                  rng,
	}
}

func GetStatisticalResources(returns [][]float64, distType int, df float64) (*StatisticalResources, error) {
	var err error

	sm := &StatisticalResources{
		DistType: distType,
		Df:       df,
	}

	sm.CovMatrix = GetCovarianceMatrix(returns)
	sm.CholeskyL, err = GetCholeskyDecomposition(sm.CovMatrix)
	if err != nil {
		return nil, err
	}

	sm.MeanReturns = make([]float64, len(returns))
	sm.StandardDeviation = make([]float64, len(returns))
	for i := range returns {
		sm.MeanReturns[i] = stat.Mean(returns[i], nil)
		sm.StandardDeviation[i] = stat.StdDev(returns[i], nil)
	}

	if distType == StudentT {
		sm.CorrMatrix = GetCorrelationMatrix(sm.CovMatrix, sm.StandardDeviation)
		sm.CholeskyCorrL, err = GetCholeskyDecomposition(sm.CorrMatrix)
		if err != nil {
			return nil, fmt.Errorf("failed to compute correlation Cholesky: %w", err)
		}
	}

	return sm, nil
}

// GetCorrelatedReturns generates one set of correlated returns
// This is goroutine-safe as long as each goroutine has its own WorkerResources
func (wr *WorkerResources) GetCorrelatedReturns() []float64 {
	switch wr.DistType {
	case StandardNormal:
		return wr.generateNormalReturns()
	case StudentT:
		return wr.generateTReturns()
	default:
		return nil
	}
}

// generateNormalReturns generates correlated normal returns
func (wr *WorkerResources) generateNormalReturns() []float64 {
	n := len(wr.MeanReturns)
	normalDist := distuv.Normal{Mu: 0, Sigma: 1, Src: wr.rng}

	z := make([]float64, n)
	for i := range n {
		z[i] = normalDist.Rand()
	}

	zVec := mat.NewVecDense(n, z)
	yVec := mat.NewVecDense(n, nil)
	yVec.MulVec(wr.CholeskyL, zVec)

	correlatedReturns := make([]float64, n) // TODO: update given understandings in test for annualization and drift
	for i := range n {
		correlatedReturns[i] = yVec.AtVec(i) + wr.MeanReturns[i]
	}

	return correlatedReturns
}

// generateTReturns generates correlated Student's t returns using Gaussian copula
func (wr *WorkerResources) generateTReturns() []float64 {
	n := len(wr.MeanReturns)
	normalDist := distuv.Normal{Mu: 0, Sigma: 1, Src: wr.rng}
	tDist := distuv.StudentsT{Mu: 0, Sigma: 1, Nu: wr.Df, Src: wr.rng}

	// generate correlated standard normals using correlation matrix
	z := make([]float64, n)
	for i := range n {
		z[i] = normalDist.Rand()
	}

	zVec := mat.NewVecDense(n, z)
	correlatedZ := mat.NewVecDense(n, nil)
	correlatedZ.MulVec(wr.CholeskyCorrL, zVec) // correlated z = chol corr * rng variables

	// gaussian copula transformation
	// https://colab.research.google.com/github/tensorflow/probability/blob/main/tensorflow_probability/examples/jupyter_notebooks/Gaussian_Copula.ipynb#scrollTo=1kSHqIp0GaRh
	correlatedReturns := make([]float64, n)
	for i := range n {
		// Transform to uniform [0,1]
		u := normalDist.CDF(correlatedZ.AtVec(i))

		// Transform to t-distributed
		tValue := tDist.Quantile(u)

		// Scale by individual sigma and add mean
		correlatedReturns[i] = tValue*wr.StandardDeviation[i] + wr.MeanReturns[i]
	}

	return correlatedReturns
}

func GetCovarianceMatrix[T ex.Number](data [][]T) *mat.SymDense {
	returnMatrix := ArrToMatrix(data)
	covMatrix := mat.NewSymDense(len(data), nil)
	stat.CovarianceMatrix(covMatrix, returnMatrix, nil)
	return covMatrix
}

func GetCorrelationMatrix(covMatrix *mat.SymDense, sigma []float64) *mat.SymDense {
	n := len(sigma)
	corrMatrix := mat.NewSymDense(n, nil)

	for i := range n {
		for j := range i + 1 {
			corr := covMatrix.At(i, j) / (sigma[i] * sigma[j])
			corrMatrix.SetSym(i, j, corr)
		}
	}

	return corrMatrix
}

func GetCholeskyDecomposition(covMatrix *mat.SymDense) (*mat.TriDense, error) {
	chol := new(mat.Cholesky)
	if ok := chol.Factorize(covMatrix); !ok {
		return nil, fmt.Errorf("covariance matrix is not positive definite")
	}

	L := new(mat.TriDense)
	chol.LTo(L)

	return L, nil
}

func ArrToMatrix[T ex.Number](data [][]T) *mat.Dense {
	nSymbols := len(data)
	nObservations := len(data[0])
	res := mat.NewDense(nObservations, nSymbols, nil)
	for j, col := range data {
		for i, row := range col {
			res.Set(i, j, float64(row))
		}
	}
	return res
}

func DotProduct[T ex.Number](a, b []T) (res T, err error) {
	if len(a) != len(b) {
		return res, fmt.Errorf("error in dotproduct, lengths of vectors are not equal")
	}

	for i, v := range a {
		res += v * b[i]
	}

	return
}
