package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Measurement struct {
	Min   float64
	Max   float64
	Sum   float64
	Count int64
}

func main() {
	measurements, err := os.Open("measurements.txt")
	if err != nil {
		panic(err)
	}
	defer measurements.Close()

	dados := make(map[string]Measurement)

	scanner := bufio.NewScanner(measurements)
	for scanner.Scan() {
		rawData := scanner.Text()
		semicolon := strings.Index(rawData, ";")
		location := rawData[:semicolon]
		rawTemp := rawData[semicolon+1:]
		temp, _ := strconv.ParseFloat(rawTemp, 64)

		measurement, ok := dados[location]
		if !ok {
			measurement = Measurement{Min: temp, Max: temp}
		}
		measurement.Min = min(measurement.Min, temp)
		measurement.Max = max(measurement.Max, temp)
		measurement.Sum += temp
		measurement.Count++

		dados[location] = measurement
	}

	locations := make([]string, 0, len(dados))
	for name := range dados {
		locations = append(locations, name)
	}

	sort.Strings(locations)

	for _, location := range locations {
		measurement := dados[location]
		fmt.Printf("%s: %#+v\n", location, measurement)
	}
}
