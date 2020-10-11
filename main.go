package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	FeeRate        = 0.003
	BenchmarkBatch = 1000

	// the original Uniswap mechanism, keep fee in liquidity pool
	FeeModelOriginal = iota
	// the same to original Uniswap, charge fee before trade, and keep fee in a separate pool
	FeeModelSeparate
	// the same to original Uniswap, charge fee after trade, and keep fee in a separate pool
	FeeModelLaterSeparate
)

type Swap struct {
	Model  int
	X, Y   float64
	IX, IY float64
	FX, FY float64
}

func NewSwap(x, y float64, model int) *Swap {
	return &Swap{
		Model: model,
		IX:    x,
		IY:    y,
		X:     x,
		Y:     y,
	}
}

func (s *Swap) Trade(amount float64) {
	k := s.X * s.Y
	switch s.Model {
	case FeeModelOriginal:
		if amount > 0 {
			fee := amount * FeeRate
			amount = amount - fee
			s.X = s.X + amount
			s.Y = k / s.X
			s.X = s.X + fee
		} else {
			fee := amount * FeeRate
			amount, fee = -amount, -fee
			amount = amount - fee
			s.Y = s.Y + amount
			s.X = k / s.Y
			s.Y = s.Y + fee
		}
	case FeeModelSeparate:
		if amount > 0 {
			fee := amount * FeeRate
			amount = amount - fee
			s.X = s.X + amount
			s.Y = k / s.X
			s.FX = s.FX + fee
		} else {
			fee := amount * FeeRate
			amount, fee = -amount, -fee
			amount = amount - fee
			s.Y = s.Y + amount
			s.X = k / s.Y
			s.FY = s.FY + fee
		}
	case FeeModelLaterSeparate:
		if amount > 0 {
			s.X = s.X + amount
			y := s.Y - k/s.X
			fee := y * FeeRate
			s.Y = s.Y - y
			s.FY = s.FY + fee
		} else {
			s.Y = s.Y - amount
			x := s.X - k/s.Y
			fee := x * FeeRate
			s.X = s.X - x
			s.FX = s.FX + fee
		}
	}
}

func (s *Swap) Simulate(trades []float64, verbose bool) (float64, float64) {
	for _, amount := range trades {
		s.Trade(amount)
	}

	dX := (s.X + s.FX - s.IX) / s.IX
	dY := (s.Y + s.FY - s.IY) / s.IY
	if verbose {
		fmt.Printf("X: %f %f%%\nY: %f %f%%\n",
			s.X+s.FX, dX*100, s.Y+s.FY, dY*100)
	}
	return dX, dY
}

func Benchmark(x, y float64, tradesGroup [][]float64, model int) {
	if model == FeeModelLaterSeparate {
		fmt.Printf("Bechmark later separate...\n")
	} else {
		fmt.Printf("Bechmark separate...\n")
	}
	var winX, winY, winXY, failXY int
	for _, trades := range tradesGroup {
		os := NewSwap(x, y, FeeModelOriginal)
		ss := NewSwap(x, y, model)
		osdX, osdY := os.Simulate(trades, false)
		ssdX, ssdY := ss.Simulate(trades, false)
		if ssdX >= osdX {
			winX += 1
		}
		if ssdY >= osdY {
			winY += 1
		}
		if ssdX >= osdX && ssdY >= osdY {
			winXY += 1
		}
		if ssdX < osdX && ssdY < osdY {
			failXY += 1
		}
	}
	fmt.Printf("X WIN: %d\nY WIN: %d\nX Y WIN: %d\nX Y FAIL: %d\n",
		winX, winY, winXY, failXY)
}

func ThresholdFullRandom(s *Swap, threshold int) []float64 {
	rand.Seed(time.Now().UnixNano())
	trades := make([]float64, 10000)
	for i := range trades {
		x := rand.Intn(int(s.IX) / threshold)
		y := -rand.Intn(int(s.IY) / threshold)
		amount := []int{x, y}[rand.Intn(2)]
		trades[i] = float64(amount)
	}
	return trades
}

func ThresholdRoundRobinRandom(s *Swap, threshold int) []float64 {
	rand.Seed(time.Now().UnixNano())
	trades := make([]float64, 10000)
	for i := range trades {
		if i%2 == 0 {
			trades[i] = float64(rand.Intn(int(s.IX) / threshold))
		} else {
			trades[i] = -float64(rand.Intn(int(s.IY) / threshold))
		}
	}
	return trades
}

func main() {
	fmt.Printf("1/3 THRESHOLD FULL RANDOM\n")
	os := NewSwap(10000, 100000, FeeModelOriginal)
	ss := NewSwap(os.IX, os.IY, FeeModelLaterSeparate)
	trades := ThresholdFullRandom(os, 3)
	fmt.Printf("Do original swap simulation...\n")
	os.Simulate(trades, true)
	fmt.Printf("Do later separate swap simulation...\n")
	ss.Simulate(trades, true)
	tradesGroup := make([][]float64, BenchmarkBatch)
	for i := range tradesGroup {
		tradesGroup[i] = ThresholdFullRandom(os, 3)
	}
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelLaterSeparate)
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelSeparate)
	fmt.Printf("\n")

	fmt.Printf("1/10 THRESHOLD FULL RANDOM\n")
	os = NewSwap(10000, 100000, FeeModelOriginal)
	ss = NewSwap(os.IX, os.IY, FeeModelLaterSeparate)
	trades = ThresholdFullRandom(os, 10)
	fmt.Printf("Do original swap simulation...\n")
	os.Simulate(trades, true)
	fmt.Printf("Do later separate swap simulation...\n")
	ss.Simulate(trades, true)
	tradesGroup = make([][]float64, BenchmarkBatch)
	for i := range tradesGroup {
		tradesGroup[i] = ThresholdFullRandom(os, 10)
	}
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelLaterSeparate)
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelSeparate)
	fmt.Printf("\n")

	fmt.Printf("1/100 THRESHOLD FULL RANDOM\n")
	os = NewSwap(10000, 100000, FeeModelOriginal)
	ss = NewSwap(os.IX, os.IY, FeeModelLaterSeparate)
	trades = ThresholdFullRandom(os, 100)
	fmt.Printf("Do original swap simulation...\n")
	os.Simulate(trades, true)
	fmt.Printf("Do later separate swap simulation...\n")
	ss.Simulate(trades, true)
	tradesGroup = make([][]float64, BenchmarkBatch)
	for i := range tradesGroup {
		tradesGroup[i] = ThresholdFullRandom(os, 100)
	}
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelLaterSeparate)
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelSeparate)
	fmt.Printf("\n")

	fmt.Printf("1/10 THRESHOLD ROUND ROBIN RANDOM\n")
	os = NewSwap(10000, 100000, FeeModelLaterSeparate)
	ss = NewSwap(os.IX, os.IY, FeeModelOriginal)
	trades = ThresholdRoundRobinRandom(os, 10)
	fmt.Printf("Do original swap simulation...\n")
	os.Simulate(trades, true)
	fmt.Printf("Do later separate swap simulation...\n")
	ss.Simulate(trades, true)
	tradesGroup = make([][]float64, BenchmarkBatch)
	for i := range tradesGroup {
		tradesGroup[i] = ThresholdRoundRobinRandom(os, 10)
	}
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelLaterSeparate)
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelSeparate)
	fmt.Printf("\n")

	fmt.Printf("1/100 THRESHOLD ROUND ROBIN RANDOM\n")
	os = NewSwap(10000, 100000, FeeModelOriginal)
	ss = NewSwap(os.IX, os.IY, FeeModelLaterSeparate)
	trades = ThresholdRoundRobinRandom(os, 100)
	fmt.Printf("Do original swap simulation...\n")
	os.Simulate(trades, true)
	fmt.Printf("Do later separate swap simulation...\n")
	ss.Simulate(trades, true)
	tradesGroup = make([][]float64, BenchmarkBatch)
	for i := range tradesGroup {
		tradesGroup[i] = ThresholdRoundRobinRandom(os, 100)
	}
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelLaterSeparate)
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelSeparate)
	fmt.Printf("\n")

	fmt.Printf("1/1000 X MONO INCREASE\n")
	os = NewSwap(10000, 100000, FeeModelOriginal)
	ss = NewSwap(os.IX, os.IY, FeeModelLaterSeparate)
	for i := range trades {
		trades[i] = -os.IY / 1000
	}
	fmt.Printf("Do original swap simulation...\n")
	os.Simulate(trades, true)
	fmt.Printf("Do later separate swap simulation...\n")
	ss.Simulate(trades, true)
	tradesGroup = make([][]float64, BenchmarkBatch)
	for i := range tradesGroup {
		tradesGroup[i] = trades
	}
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelLaterSeparate)
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelSeparate)
	fmt.Printf("\n")

	fmt.Printf("1/1000 X Y HALF INCREASE\n")
	os = NewSwap(10000, 100000, FeeModelOriginal)
	ss = NewSwap(os.IX, os.IY, FeeModelLaterSeparate)
	for i := range trades {
		if i < len(trades)/2 {
			trades[i] = os.IX / 1000
		} else {
			trades[i] = -os.IY / 1000
		}
	}
	fmt.Printf("Do original swap simulation...\n")
	os.Simulate(trades, true)
	fmt.Printf("Do later separate swap simulation...\n")
	ss.Simulate(trades, true)
	tradesGroup = make([][]float64, BenchmarkBatch)
	for i := range tradesGroup {
		tradesGroup[i] = trades
	}
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelLaterSeparate)
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelSeparate)
	fmt.Printf("\n")

	fmt.Printf("1/100 X Y ALWAYS DRAW\n")
	os = NewSwap(10000, 100000, FeeModelOriginal)
	ss = NewSwap(os.IX, os.IY, FeeModelLaterSeparate)
	for i := range trades {
		if i%2 == 0 {
			trades[i] = os.IX / 100
		} else {
			trades[i] = -os.IY / 100
		}
	}
	fmt.Printf("Do original swap simulation...\n")
	os.Simulate(trades, true)
	fmt.Printf("Do later separate swap simulation...\n")
	ss.Simulate(trades, true)
	tradesGroup = make([][]float64, BenchmarkBatch)
	for i := range tradesGroup {
		tradesGroup[i] = trades
	}
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelLaterSeparate)
	Benchmark(os.IX, os.IY, tradesGroup, FeeModelSeparate)
	fmt.Printf("\n")
}
