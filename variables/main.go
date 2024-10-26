package main

import (
	"fmt"
	"strconv"
)

// bool
// int int8 int16 int32 int64
// uint uint8 uint16 uint32 uint64 uintptr // inteiros apenas positivos
// byte // alias para uint8
// rune // alias para int32 // representa um code point
// float32 float64
// complex64 complex128
// string
func main() {
	identifyExecution("variables", variables)
	identifyExecution("pointers", pointers)
	identifyExecution("slice", slice)
	identifyExecution("appendInSlice", appendInSlice)
	identifyExecution("sliceBoundChecks", sliceBoundChecks)
}

func identifyExecution(identifier string, function func()) {
	fmt.Println("----- START ", identifier, " -----")
	function()
	fmt.Println("----- END ", identifier, " -----")
}
func variables() {
	var i int = 10084
	// conversão de int para float64
	f := float64(i)
	fmt.Println(i, f)

	r := string(i) // conversão de int para rune de ASCII
	fmt.Println(r)

	s := strconv.FormatInt(int64(i), 10) // conversão de int para string
	fmt.Println(s)

	// constantes podem ser apenas strings, runes, inteiros, floats e complexos
	const c = 10 // se não for atribuído um tipo, o tipo é inferido mas se converte para o contexto necessário
	takeInt32(c)
	takeInt64(c)

	// const c2 int = 20
	// takeInt32(c2) // não compila da erro por ter tipo específico

	const t int = 10
	var t2 int = t
	fmt.Println(t2)

	arrEmpty := [3]int{}
	arr := [3]int{1, 2, 3}
	arrIndex := [5]int{1: 1, 3: 2}
	fmt.Println(arrEmpty)
	fmt.Println(arr)
	fmt.Println(arrIndex)
}

func takeInt32(i int32) {
	fmt.Println(i)
}
func takeInt64(i int64) {
	fmt.Println(i)
}

func pointers() {
	var a int = 10
	var b *int = &a
	fmt.Println(a, b, *b)

	x := 10
	y := 10
	z := takeX(x)
	takeY(&y)
	fmt.Println(x, z, y)

}

func takeX(h int) int {
	fmt.Println("TakeX", h)
	h = 100
	return h
}

func takeY(y *int) {
	*y = 100
}

func slice() {
	arr := [5]int{1, 2, 3, 4, 5}
	slice := arr[1:3]
	fmt.Println("SLICES", slice)
}

func appendInSlice() {
	// array -> because has fixed size
	moviesList := [5]string{
		"Titanic",
		"The Godfather",
		"The Godfather II",
		"The Godfather III",
		"The Godfather IV",
	}
	fmt.Println(moviesList)
	// slice -> because has dynamic size, if not have more capacity, it will double the actual capacity
	var movies []string
	fmt.Println(len(movies), cap(movies), movies)
	movies = append(movies, "Titanic")
	fmt.Println(len(movies), cap(movies), movies)
	movies = append(movies, "The Godfather")
	fmt.Println(len(movies), cap(movies), movies)
	movies = append(movies, "The Godfather II")
	fmt.Println(len(movies), cap(movies), movies)
	movies = append(movies, "The Godfather III")
	fmt.Println(len(movies), cap(movies), movies)
	movies = append(movies, "The Godfather IV")
	fmt.Println(len(movies), cap(movies), movies)

	var movies2 = make([]string, 0, len(moviesList))
	fmt.Println(len(movies2), cap(movies2), movies2)
}

func sliceBoundChecks() {
	slice := []int{1, 2, 3, 4, 5}
	_ = slice[4] // bound check to not degrade performance,
	// withou bound check, the program will check if the index is in the slice every call
	fmt.Println(slice[0])
	fmt.Println(slice[1])
	fmt.Println(slice[2])
	fmt.Println(slice[3])
	fmt.Println(slice[4])
}
