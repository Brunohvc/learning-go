package main

import (
	"fmt"
	"functions/pacote"
	"os"
)

var idade int

func main() {
	const nome string = "Gopher"
	sobrenome := "Silva"
	fmt.Println("Hello, World!", nome, sobrenome, idade)
	fmt.Println(pacote.Foo)
	fmt.Println(pacote.Bar)
	pacote.PrintMinha()
	fmt.Println("Somar 1 + 2 = ", somar(1, 2))
	fmt.Println("Subtrair 2 - 1 = ", subtrair(2, 1))
	a, b := swap(1, 2)
	fmt.Println("Swap 1, 2 = ", a, b)
	res, rem := dividir(5, 2)
	fmt.Println("Dividir 5 / 2 = ", res, "resto", rem)
	mult := multiplicar(2)
	fmt.Println("Multiplicar 2 * 3 = ", mult(3))
	fmt.Println("Somar variadico 1 + 2 + 3 + 4 = ", somarVariadico(1, 2, 3, 4))
	doDefer()
	doDeferLastInFirstOut()
	openFile()
}

func somar(a int, b int) int {
	return a + b
}

func subtrair(a, b int) int {
	return a - b
}

func swap(a, b int) (int, int) {
	return b, a
}

func dividir(a, b int) (res int, rem int) {
	res = a / b
	rem = a % b
	return res, rem
}

func multiplicar(a int) func(int) int {
	return func(b int) int {
		return a * b
	}
}

func somarVariadico(args ...int) int {
	total := 0
	for _, v := range args {
		total += v
	}
	return total
}

func doDefer() {
	defer fmt.Println("DEFER world")
	fmt.Println("DEFER hello")
}

func doDeferLastInFirstOut() {
	defer fmt.Println("DEFER 3")
	defer fmt.Println("DEFER 2")
	fmt.Println("DEFER 1")
}

func openFile() {
	f, err := os.Open("./file.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	// print file content
	b := make([]byte, 100)
	n, err := f.Read(b)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("string(b):", string(b))
	fmt.Println("n:", n)
	fmt.Println("b[:n]:", b[:n])
	fmt.Println("string(b[:n])", string(b[:n]))
}
