package main

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestCalculations(t *testing.T) {
	// single number
	testCalc(t, []byte("1"), 1)
	testCalc(t, []byte("69"), 69)
	testCalc(t, []byte(".420"), .420)
	testCalc(t, []byte("33.33"), 33.33)
	testCalc(t, []byte("1337."), 1337)

	// operations
	testCalc(t, []byte("3+4"), 7)
	testCalc(t, []byte("2-5"), -3)
	testCalc(t, []byte("3*2"), 6)
	testCalc(t, []byte("3(-2)"), -6)
	testCalc(t, []byte("(3)2"), 6)
	testCalc(t, []byte("(3)(-2)"), -6)
	testCalc(t, []byte("3/2"), 1.5)
	testCalc(t, []byte("4//6"), 0)
	testCalc(t, []byte("7%10"), 7)
	testCalc(t, []byte("3v27"), 3)
	testCalc(t, []byte("2v-2"), math.NaN())
	testCalc(t, []byte("(1+2)v27"), 3)
	testCalc(t, []byte("2**8"), 256)

	// associative property
	testCalc(t, []byte("3+4+5+6"), 18)
	testCalc(t, []byte("3+4-5+6"), 8)
	testCalc(t, []byte("3-4+5+6"), 10)
	testCalc(t, []byte("3-4-5+6"), 0)

	testCalc(t, []byte("3*4*5*2"), 120)
	testCalc(t, []byte("3*4/5*2"), 4.8)
	testCalc(t, []byte("3/4*5*2"), 7.5)
	testCalc(t, []byte("3/4/5*2"), 0.3)

	// operators precedence
	testCalc(t, []byte("3+4*5"), 23)
	testCalc(t, []byte("3*4+5"), 17)
	testCalc(t, []byte("3*4v5"), 4.486046343663661)
	testCalc(t, []byte("3-4**5"), -1021)

	// brackets
	testCalc(t, []byte("2+(3)"), 5)
	testCalc(t, []byte("(2)+(3)"), 5)
	testCalc(t, []byte("(2+3)"), 5)
	testCalc(t, []byte("((2)+(3))"), 5)
	testCalc(t, []byte("(2)+(3)4(2+3-1)"), 50)
	testCalc(t, []byte("(((2(2"), 4)

	// space expansion
	testCalc(t, []byte("1+ 1"), 2)
	testCalc(t, []byte("1 + 1 + 1 * 2"), 4)
	testCalc(t, []byte("1+1 *2"), 4)
	testCalc(t, []byte("( 1+1*2)"), 3)
	testCalc(t, []byte("(1+1 *2"), 4)
	testCalc(t, []byte("1+1 *2 **2"), 16)
	testCalc(t, []byte("1+1 *2** 2"), 8)
}

func TestInvalidSyntax(t *testing.T) {
	assertCalcError(t, []byte("1("))
	assertCalcError(t, []byte(")1"))
	assertCalcError(t, []byte("1+"))
	assertCalcError(t, []byte("+1"))
	assertCalcError(t, []byte("1+1++1"))
	assertCalcError(t, []byte("1+1()"))
	assertCalcError(t, []byte("1+1 1"))
	assertCalcError(t, []byte("1+1+"))
	assertCalcError(t, []byte("1+1("))
	assertCalcError(t, []byte("1+1)"))
}

func TestFuzzyInput(t *testing.T) {
	seed := rand.Int63()
	// seed = 2800921359699683501
	fmt.Println("seed:", seed)
	random := rand.New(rand.NewSource(seed))

	t.Run("good tokens", func(t *testing.T) {
		for range 1_000_000 {
			input := createRandomInput(random)
			Calc(input)
		}
	})

	t.Run("bad tokens", func(t *testing.T) {
		for range 100_000 {
			input := createRandomBadInput(random)
			Calc(input)
		}
	})

	t.Run("good input", func(t *testing.T) {
		for range 1_000_000 {
			input := createRandomGoodInput(random)
			_, _, err := Calc(input)
			if err != nil && err.Error() != "eval tree: division by 0" {
				t.Errorf("error: input = %q: %v", input, err)
				return
			}
		}
	})
}

func createRandomGoodInput(random *rand.Rand) []byte {
	n := random.Int()%100 + 1
	digits := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	operators := []string{"+", "-", "*", "/", "//", "%", "v", "**"}
	input := make([]byte, 0, n+8)

	for len(input) < n {
		if random.Intn(15) == 0 {
			input = append(input, ' ')
		}
		if random.Intn(15) == 0 {
			input = append(input, '(')
		}

		var digit byte
		for {
			random.Shuffle(len(digits), func(i, j int) {
				digits[i], digits[j] = digits[j], digits[i]
			})
			digit = digits[0]
			if len(input) > 0 && (input[len(input)-1] == '/' || input[len(input)-1] == '%') && digit == '0' {
				continue
			}
			break
		}
		input = append(input, digit)

		random.Shuffle(len(operators), func(i, j int) {
			operators[i], operators[j] = operators[j], operators[i]
		})
		input = append(input, operators[0]...)
	}

	for _, op := range operators {
		for bytes.HasSuffix(input, []byte(op)) {
			input = input[:len(input)-len(op)]
		}
	}

	return input
}

func createRandomInput(random *rand.Rand) []byte {
	n := random.Int() % 100
	tokens := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ".", "+", "-", "*", "/", "//", "%", "v", "**", "(", ")", " "}
	input := make([]byte, 0, n+8)

	for len(input) < n {
		random.Shuffle(len(tokens), func(i, j int) {
			tokens[i], tokens[j] = tokens[j], tokens[i]
		})
		input = append(input, tokens[0]...)
	}

	return input
}

func createRandomBadInput(random *rand.Rand) []byte {
	n := random.Int() % 1000
	input := make([]byte, 0, n+8)

	for len(input) < n {
		input = append(input, byte(random.Int()%256))
	}

	return input
}

func testCalc(t *testing.T, input []byte, expected float64) {
	errMargin := 0.0000000001
	res, processed, err := Calc(input)
	if err != nil {
		t.Errorf("error: input = %q, processed = %q: %v", input, processed, err)
	}
	if res < expected-errMargin || res > expected+errMargin {
		t.Errorf("calculation error: input = %q, processed = %q: expected %.20f, got %.20f", input, processed, expected, res)
	}
}

func assertCalcError(t *testing.T, input []byte) {
	_, _, err := Calc(input)
	if err == nil {
		t.Errorf("error expected: input = %q", input)
	}
}
