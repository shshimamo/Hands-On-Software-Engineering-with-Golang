package fizzbuzz

import "fmt"

func Evaluate(n int) string {
	if n != 0 {
		switch {
		case n%3 == 0 && n%5 == 0:
			return "FizzBuzz"
		case n%3 == 0:
			return "Fizz"
		case n%5 == 0:
			return "Buzz"
		}
	}
	return fmt.Sprint(n)
}
