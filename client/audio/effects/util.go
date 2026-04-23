package effects

import "math"

func Amp2Db(amp float64) float64 {
	return 20 * math.Log(amp) / math.Ln10
}

func Db2Amp(db float64) float64 {
	return math.Pow(10, db/20)
}