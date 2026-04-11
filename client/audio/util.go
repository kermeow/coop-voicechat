package audio

import "math"

func amp2db(amp float64) float64 {
	return 20 * math.Log(amp) / math.Ln10
}

func db2amp(db float64) float64 {
	return math.Pow(10, db/20)
}
