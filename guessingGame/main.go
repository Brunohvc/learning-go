package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Welcome to the guessing game!")
	fmt.Println("Please think of a number between 1 and 100")

	response := rand.Int64N(101)
	scanner := bufio.NewScanner(os.Stdin)
	guess := [10]int64{}

	for index := range guess {
		fmt.Println("What is your guess?")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		guessInt, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			fmt.Println("Invalid input. Please enter an integer.")
			return
		}
		guess[index] = guessInt

		switch {
		case guessInt < response:
			fmt.Println("Your guess is too low: ", guessInt)
		case guessInt > response:
			fmt.Println("Your guess is too high: ", guessInt)
		default:
			fmt.Printf(
				"Congratulations! You guessed the number %d\n"+
					"Your guess count: %d\n"+
					"You had the following guesses: %v\n",
				response, index+1, guess[:index+1],
			)
			return
		}
	}

	fmt.Printf(
		"Sorry, you didn't guess the number. It was: %d\n"+
			"You had the following guesses: %v\n",
		response, guess,
	)
}
