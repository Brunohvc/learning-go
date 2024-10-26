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

	pointers()
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
