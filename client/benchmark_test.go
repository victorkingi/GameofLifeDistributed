package main

import (
	"fmt"
	"os"
	"testing"
	"uk.ac.bris.cs/gameoflife/gol"
)

func numberOfThreads() []int {
	return []int{1} //5120 seems to cause terminal to hang
}

func Benchmark16(b *testing.B) {
	for _, threadNo := range numberOfThreads() {
		b.Run(fmt.Sprint(threadNo), func(b *testing.B) {
			os.Stdout = nil // Disable all program output apart from benchmark results
			p := gol.Params{
				Turns:       10000,
				Threads:     threadNo,
				ImageWidth:  16,
				ImageHeight: 16,
			}
			for i := 0; i < b.N; i++ {
				b.StartTimer()
				events := make(chan gol.Event)
				gol.Run(p, events, nil)
				for event := range events {
					switch e := event.(type) {
					case gol.FinalTurnComplete:
						fmt.Println(e)
					}
				}
				b.StopTimer()
			}
		})
	}
}

func Benchmark64(b *testing.B) {
	for _, threadNo := range numberOfThreads() {
		b.Run(fmt.Sprint(threadNo), func(b *testing.B) {
			os.Stdout = nil // Disable all program output apart from benchmark results
			p := gol.Params{
				Turns:       10000,
				Threads:     threadNo,
				ImageWidth:  64,
				ImageHeight: 64,
			}
			for i := 0; i < b.N; i++ {
				b.StartTimer()
				events := make(chan gol.Event)
				gol.Run(p, events, nil)
				for event := range events {
					switch e := event.(type) {
					case gol.FinalTurnComplete:
						fmt.Println(e)
					}
				}
				b.StopTimer()
			}
		})
	}
}

func Benchmark128(b *testing.B) {
	for _, threadNo := range numberOfThreads() {
		b.Run(fmt.Sprint(threadNo), func(b *testing.B) {
			os.Stdout = nil // Disable all program output apart from benchmark results
			p := gol.Params{
				Turns:       10000,
				Threads:     threadNo,
				ImageWidth:  128,
				ImageHeight: 128,
			}
			for i := 0; i < b.N; i++ {
				b.StartTimer()
				events := make(chan gol.Event)
				gol.Run(p, events, nil)
				for event := range events {
					switch e := event.(type) {
					case gol.FinalTurnComplete:
						fmt.Println(e)
					}
				}
				b.StopTimer()
			}
		})
	}
}

func Benchmark256(b *testing.B) {
	for _, threadNo := range numberOfThreads() {
		b.Run(fmt.Sprint(threadNo), func(b *testing.B) {
			os.Stdout = nil // Disable all program output apart from benchmark results
			p := gol.Params{
				Turns:       10000,
				Threads:     threadNo,
				ImageWidth:  256,
				ImageHeight: 256,
			}
			for i := 0; i < b.N; i++ {
				b.StartTimer()
				events := make(chan gol.Event)
				gol.Run(p, events, nil)
				for event := range events {
					switch e := event.(type) {
					case gol.FinalTurnComplete:
						fmt.Println(e)
					}
				}
				b.StopTimer()
			}
		})
	}
}

func Benchmark512(b *testing.B) {
	for _, threadNo := range numberOfThreads() {
		b.Run(fmt.Sprint(threadNo), func(b *testing.B) {
			os.Stdout = nil // Disable all program output apart from benchmark results
			p := gol.Params{
				Turns:       10000,
				Threads:     threadNo,
				ImageWidth:  512,
				ImageHeight: 512,
			}
			for i := 0; i < b.N; i++ {
				b.StartTimer()
				events := make(chan gol.Event)
				gol.Run(p, events, nil)
				for event := range events {
					switch e := event.(type) {
					case gol.FinalTurnComplete:
						fmt.Println(e)
					}
				}
				b.StopTimer()
			}
		})
	}
}

func Benchmark5120(b *testing.B) {
	for _, threadNo := range numberOfThreads() {
		b.Run(fmt.Sprint(threadNo), func(b *testing.B) {
			os.Stdout = nil // Disable all program output apart from benchmark results
			p := gol.Params{
				Turns:       10,
				Threads:     threadNo,
				ImageWidth:  5120,
				ImageHeight: 5120,
			}
			for i := 0; i < b.N; i++ {
				b.StartTimer()
				events := make(chan gol.Event)
				gol.Run(p, events, nil)
				for event := range events {
					switch e := event.(type) {
					case gol.FinalTurnComplete:
						fmt.Println(e)
					}
				}
				b.StopTimer()
			}
		})
	}
}
