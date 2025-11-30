package core

import (
	"fmt"

	"golang.org/x/exp/constraints"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

const (
	StandardNormal = iota
	StudentT
)

type StatisticalResources struct {
	CovMatrix             *mat.SymDense
	CholeskyDecomposition *mat.TriDense
	Mu                    []float64
	Sigma                 []float64
	TypeOfDistribution    int
	Distributions         []Distribution
	Df                    int
}

type Distribution interface {
	Rand() float64
}

type Number interface {
	constraints.Integer | constraints.Float
}

func GetStatisticalResources(returns [][]float64, typeOfDistribution int, seeds []int) (res StatisticalResources, err error) {
	nSymbols := len(returns)
	nObservations := len(returns[0])

	res.CovMatrix = GetCovarianceMatrix(returns)
	res.CholeskyDecomposition, err = GetCholeskyDecomposition(res.CovMatrix)
	if err != nil {
		return
	}

	mu := make([]float64, len(returns))
	sigma := make([]float64, len(returns))
	for i := range returns {
		mu[i] = stat.Mean(returns[i], nil)
		sigma[i] = stat.StdDev(returns[i], nil)
	}

	distributions := make([]Distribution, len(seeds))
	for i, v := range seeds {

	}

	return
}

func GetDistribution(typeOfDistribution, seed int) Distribution {
	switch typeOfDistribution {
	case StandardNormal:
		return nil
	case StudentT:
		return nil
	default:
		return nil
	}
}

func GetCorrelatedReturnsFromCovariance() []float64 {

}

func GetCorrelatedReturnsFromCorrelationMatric() []float64 {
	
}

func GetCovarianceMatrix[T Number](data [][]T) *mat.SymDense {
	returnMatrix := ArrToMatrix(data)
	covMatrix := mat.NewSymDense(len(data), nil)
	stat.CovarianceMatrix(covMatrix, returnMatrix, nil)
	return covMatrix
}

func GetCholeskyDecomposition(covMatrix *mat.SymDense) (*mat.TriDense, error) {
	var chol mat.Cholesky
	if ok := chol.Factorize(covMatrix); !ok {
		return nil, fmt.Errorf("covariance matrix is not positive definite")
	}

	L := new(mat.TriDense)
	chol.LTo(L)

	return L, nil
}

func ArrToMatrix[T Number](data [][]T) *mat.Dense {
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

func DotProduct[T Number](a, b []T) (res T, err error) {
	if len(a) != len(b) {
		return res, fmt.Errorf("error in dotproduct, lengths of vectors are not equal")
	}

	for i, v := range a {
		res += v * b[i]
	}

	return
}
