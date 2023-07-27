package main

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/stat"
)

// resultXY --> sum((x-meanX)*(y-meanY))
// resultXX --> sum((x-meanX)^2)

func sumXYandXX(arrayX []float64, arrayY []float64, meanX float64, meanY float64) (float64, float64) {
	resultXX := 0.0
	resultXY := 0.0
	for x := 0; x < len(arrayX); x++ {
		for y := 0; y < len(arrayY); y++ {
			if x == y {
				resultXY += (arrayX[x] - meanX) * (arrayY[y] - meanY)
			}
		}
		resultXX += (arrayX[x] - meanX) * (arrayX[x] - meanX)
	}
	return resultXY, resultXX
}

// estimateBoB1 --> Function that calculates the regression coefficients b0 and b1
// y_predicted = b0 + b1*x_input
func estimateB0B1(x []float64, y []float64) (float64, float64) {
	var meanX float64
	var meanY float64
	var sumXY float64
	var sumXX float64
	meanX = stat.Mean(x, nil) //mean of x
	meanY = stat.Mean(y, nil) //mean pf y
	sumXY, sumXX = sumXYandXX(x, y, meanX, meanY)
	// regression coefficients
	b1 := sumXY / sumXX    // b1 or slope
	b0 := meanY - b1*meanX // b0 or intercept
	return b0, b1
}

func rmseCost(y_predicted []float64, y_test []float64) float64 {
	sz := len(y_test)
	var rmse float64 = 0.0
	for i := 0; i < len(y_test); i++ {
		rmse = rmse + math.Abs(y_test[i]-y_predicted[i])*math.Abs(y_test[i]-y_predicted[i])
	}
	rmse = rmse / float64(sz)
	rmse = math.Sqrt(rmse)
	return rmse
}

func main() {
	var X_data = []float64{4.4, 4.1, 3.81}
	var y_data = []float64{4, 6.89, 9.18}

	b0, b1 := estimateB0B1(X_data, y_data)
	fmt.Println("Equacao =", b0, " +", b1, "*x")
}
