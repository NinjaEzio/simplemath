package main

import (
	"fmt"
	"os"
	"simplemath/client/rpc"
)

func main() {
	if len(os.Args) == 1 {
		usage()
		os.Exit(1)
	}
	method := os.Args[1]
	switch method {
	case "gcd":
		if len(os.Args) < 4 {
			usage()
			os.Exit(1)
		}
		rpc.GreatCommonDivisor(os.Args[2], os.Args[3])
	case "fibo":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		rpc.GetFibonacci(os.Args[2])
	case "stat":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		rpc.Statistics(os.Args[2])
	case "prime":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		rpc.PrimeFactorization(os.Args[2])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Welcome to Simple Math Client")
	fmt.Println("Usage:")
	fmt.Println("gcd num1 num2")
	fmt.Println("fibo num")
	fmt.Println("stat count(the number of numbers)")
	fmt.Println("prime count(the number of numbers)")
	fmt.Println("Enjoy")
}
