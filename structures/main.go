package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("INIT")
	loops()
	loopSync()
	conditions()
	today := time.Now()
	fmt.Println("Today is:", today.Weekday(), "Is weekend?", isWeekend(today))
}

func loops() {
	for i := 0; i < 2; i++ {
		// wirte number after "Hello 'number' times"
		fmt.Println(`Hello`, i, `times`)
	}

	for {
		fmt.Println("loop")
		break
	}

	arr := [3]int{5, 6, 7}

	for index, element := range arr {
		fmt.Println(index, element)
		// print memory address
		fmt.Println(&index, &element)
	}

	for range 3 {
		fmt.Println("loop")
	}
}

func loopSync() {
	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(param int) {
			fmt.Println("SYNC", param)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func conditions() {
	// if
	if 1 == 1 {
		fmt.Println("1 == 1")
	}

	// if else
	if 1 == 2 {
		fmt.Println("1 == 2")
	} else {
		fmt.Println("1 != 2")
	}

	// if else if
	if 1 == 2 {
		fmt.Println("1 == 2")
	} else if 1 == 1 {
		fmt.Println("1 == 1")
	}

	// switch
	switch 1 {
	case 1:
		fmt.Println("1")
	case 2:
		fmt.Println("2")
	default:
		fmt.Println("default")
	}

	// switch with condition
	switch {
	case 1 == 1:
		fmt.Println("1 == 1")
	case 1 == 2:
		fmt.Println("1 == 2")
	default:
		fmt.Println("default")
	}

	const x = 1
	switch x {
	case 1:
		fmt.Println("fallthrough 1")
		fallthrough
	case 2:
		fmt.Println("fallthrough 2")
	default:
		fmt.Println("fallthrough default")
	}

	if err := doError(); err != nil {
		fmt.Println("My error is", err)
	}

}

func doError() error {
	return errors.New("ERROR TEST")
}

func isWeekend(day time.Time) bool {
	switch day.Weekday() {
	case time.Saturday, time.Sunday:
		return true
	default:
		return false
	}
}
