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
	testStatement(t, nil, []byte("1"), 1)
	testStatement(t, nil, []byte("69"), 69)
	testStatement(t, nil, []byte(".420"), .420)
	testStatement(t, nil, []byte("33.33"), 33.33)
	testStatement(t, nil, []byte("1337."), 1337)

	// operations
	testStatement(t, nil, []byte("3+4"), 7)
	testStatement(t, nil, []byte("2-5"), -3)
	testStatement(t, nil, []byte("3*2"), 6)
	testStatement(t, nil, []byte("3(-2)"), -6)
	testStatement(t, nil, []byte("(3)2"), 6)
	testStatement(t, nil, []byte("(3)(-2)"), -6)
	testStatement(t, nil, []byte("3/2"), 1.5)
	testStatement(t, nil, []byte("4//6"), 0)
	testStatement(t, nil, []byte("7%10"), 7)
	testStatement(t, nil, []byte("3v27"), 3)
	testStatement(t, nil, []byte("2v-2"), math.NaN())
	testStatement(t, nil, []byte("(1+2)v27"), 3)
	testStatement(t, nil, []byte("2**8"), 256)

	// associative property
	testStatement(t, nil, []byte("3+4+5+6"), 18)
	testStatement(t, nil, []byte("3+4-5+6"), 8)
	testStatement(t, nil, []byte("3-4+5+6"), 10)
	testStatement(t, nil, []byte("3-4-5+6"), 0)

	testStatement(t, nil, []byte("3*4*5*2"), 120)
	testStatement(t, nil, []byte("3*4/5*2"), 4.8)
	testStatement(t, nil, []byte("3/4*5*2"), 7.5)
	testStatement(t, nil, []byte("3/4/5*2"), 0.3)

	// operators precedence
	testStatement(t, nil, []byte("3+4*5"), 23)
	testStatement(t, nil, []byte("3*4+5"), 17)
	testStatement(t, nil, []byte("3*4v5"), 4.486046343663661)
	testStatement(t, nil, []byte("3-4**5"), -1021)

	// brackets
	testStatement(t, nil, []byte("2+(3)"), 5)
	testStatement(t, nil, []byte("(2)+(3)"), 5)
	testStatement(t, nil, []byte("(2+3)"), 5)
	testStatement(t, nil, []byte("((2)+(3))"), 5)
	testStatement(t, nil, []byte("(2)+(3)4(2+3-1)"), 50)
	testStatement(t, nil, []byte("(((2(2"), 4)

	// space expansion
	testStatement(t, nil, []byte("1+ 1"), 2)
	testStatement(t, nil, []byte("1 + 1 + 1 * 2"), 4)
	testStatement(t, nil, []byte("1+1 *2"), 4)
	testStatement(t, nil, []byte("( 1+1*2)"), 3)
	testStatement(t, nil, []byte("(1+1 *2"), 4)
	testStatement(t, nil, []byte("1+1 *2 **2"), 16)
	testStatement(t, nil, []byte("1+1 *2** 2"), 8)
}

func TestVariables(t *testing.T) {
	vars := make(map[string]float64)

	testStatement(t, vars, []byte("A = 1+1 *2"), 4)
	testStatement(t, vars, []byte("A"), 4)
	testStatement(t, vars, []byte("A+1"), 5)
	testStatement(t, vars, []byte("A+A"), 8)
	testStatement(t, vars, []byte("B = A*A"), 16)
	testStatement(t, vars, []byte("B"), 16)
	testStatement(t, vars, []byte("snake_case_69 = B"), 16)
	testStatement(t, vars, []byte("snake_case_69 + 1"), 17)
}

func TestInvalidSyntax(t *testing.T) {
	assertStatementError(t, []byte("1("))
	assertStatementError(t, []byte(")1"))
	assertStatementError(t, []byte("1+"))
	assertStatementError(t, []byte("+1"))
	assertStatementError(t, []byte("1+1++1"))
	assertStatementError(t, []byte("1+1()"))
	assertStatementError(t, []byte("1+1 1"))
	assertStatementError(t, []byte("1+1+"))
	assertStatementError(t, []byte("1+1("))
	assertStatementError(t, []byte("1+1)"))
}

func TestFuzzyInput(t *testing.T) {
	seed := rand.Int63()
	// seed = 2800921359699683501
	fmt.Println("seed:", seed)
	random := rand.New(rand.NewSource(seed))

	m := make(map[string]float64)

	t.Run("good tokens", func(t *testing.T) {
		for range 1_000_000 {
			input := createRandomInput(random)
			EvalStatement(input, m)
		}
	})

	t.Run("bad tokens", func(t *testing.T) {
		for range 100_000 {
			input := createRandomBadInput(random)
			EvalStatement(input, m)
		}
	})

	t.Run("good input", func(t *testing.T) {
		for range 1_000_000 {
			input := createRandomGoodInput(random)
			_, _, _, err := EvalStatement(input, m)
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

	input = bytes.Trim(input, " \t\r\n")
	return input
}

func createRandomInput(random *rand.Rand) []byte {
	n := random.Int()%100 + 1
	tokens := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ".", "+", "-", "*", "/", "//", "%", "v", "**", "(", ")", " "}
	input := make([]byte, 0, n+8)

	for len(input) < n {
		random.Shuffle(len(tokens), func(i, j int) {
			tokens[i], tokens[j] = tokens[j], tokens[i]
		})
		input = append(input, tokens[0]...)
	}

	input = bytes.Trim(input, " \t\r\n")
	return input
}

func createRandomBadInput(random *rand.Rand) []byte {
	n := random.Int()%100 + 1
	input := make([]byte, 0, n+8)

	for len(input) < n {
		input = append(input, byte(random.Int()%256))
	}

	input = bytes.Trim(input, " \t\r\n")
	return input
}

func testStatement(t *testing.T, vars map[string]float64, input []byte, expected float64) {
	errMargin := 0.0000000001
	res, _, processed, err := EvalStatement(input, vars)
	if err != nil {
		t.Errorf("error: input = %q, processed = %q: %v", input, processed, err)
	}
	if res < expected-errMargin || res > expected+errMargin {
		t.Errorf("calculation error: input = %q, processed = %q: expected %.20f, got %.20f", input, processed, expected, res)
	}
}

func assertStatementError(t *testing.T, input []byte) {
	_, _, _, err := EvalStatement(input, nil)
	if err == nil {
		t.Errorf("error expected: input = %q", input)
	}
}
