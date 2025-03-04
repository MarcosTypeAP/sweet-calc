package main

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"
)

func TestCalculations(t *testing.T) {
	// single number
	testStatement(t, nil, "1", 1)
	testStatement(t, nil, "69", 69)
	testStatement(t, nil, ".420", .420)
	testStatement(t, nil, "33.33", 33.33)
	testStatement(t, nil, "1337.", 1337)
	testStatement(t, nil, "69_420", 69420)
	testStatement(t, nil, "69___420__", 69420)
	assertStatementError(t, "__69___420__")
	assertStatementError(t, "4.2_0")

	// operations
	testStatement(t, nil, "3+4", 7)
	testStatement(t, nil, "2-5", -3)
	testStatement(t, nil, "3*2", 6)
	testStatement(t, nil, "3(-2)", -6)
	testStatement(t, nil, "(3)2", 6)
	testStatement(t, nil, "(3)(-2)", -6)
	testStatement(t, nil, "3/2", 1.5)
	testStatement(t, nil, "4//6", 0)
	testStatement(t, nil, "7%10", 7)
	testStatement(t, nil, "3v27", 3)
	assertStatementError(t, "2v-2")
	testStatement(t, nil, "(1+2)v27", 3)
	testStatement(t, nil, "2**8", 256)

	// functions
	testStatement(t, nil, "sin 2", math.Sin(2))
	testStatement(t, nil, "cos 2", math.Cos(2))
	testStatement(t, nil, "tan 2", math.Tan(2))
	testStatement(t, nil, "asin .1", math.Asin(0.1))
	assertStatementError(t, "asin 2")
	testStatement(t, nil, "acos .1", math.Acos(0.1))
	assertStatementError(t, "acos 2")
	testStatement(t, nil, "atan 2", math.Atan(2))
	testStatement(t, nil, "sin -.2", math.Sin(-0.2))
	testStatement(t, nil, "sin cos 2", math.Sin(math.Cos(2)))
	testStatement(t, nil, "sin cos -2 * 2", math.Sin(math.Cos(-2))*2)
	testStatement(t, nil, "sin 4*2", math.Sin(4*2))
	testStatement(t, nil, "sin 4 *2", math.Sin(4)*2)
	testStatement(t, nil, "sin (4 *2", math.Sin(4*2))

	// associative property
	testStatement(t, nil, "3+4+5+6", 18)
	testStatement(t, nil, "3+4-5+6", 8)
	testStatement(t, nil, "3-4+5+6", 10)
	testStatement(t, nil, "3-4-5+6", 0)

	testStatement(t, nil, "3*4*5*2", 120)
	testStatement(t, nil, "3*4/5*2", 4.8)
	testStatement(t, nil, "3/4*5*2", 7.5)
	testStatement(t, nil, "3/4/5*2", 0.3)

	// operators precedence
	testStatement(t, nil, "3+4*5", 23)
	testStatement(t, nil, "3*4+5", 17)
	testStatement(t, nil, "3*4v5", 4.486046343663661)
	testStatement(t, nil, "3-4**5", -1021)

	// brackets
	testStatement(t, nil, "2+(3)", 5)
	testStatement(t, nil, "(2)+(3)", 5)
	testStatement(t, nil, "(2+3)", 5)
	testStatement(t, nil, "((2)+(3))", 5)
	testStatement(t, nil, "(2)+(3)4(2+3-1)", 50)
	testStatement(t, nil, "(((2(2", 4)

	// space expansion
	testStatement(t, nil, "1+ 1", 2)
	testStatement(t, nil, "1 + 1 + 1 * 2", 4)
	testStatement(t, nil, "1+1 *2", 4)
	testStatement(t, nil, "( 1+1*2)", 3)
	testStatement(t, nil, "(1+1 *2", 4)
	testStatement(t, nil, "1+1 *2 **2", 16)
	testStatement(t, nil, "1+1 *2** 2", 8)
}

func TestVariables(t *testing.T) {
	vars := make(map[string]float64)

	testStatement(t, vars, "A = 1+1 *2", 4)
	testStatement(t, vars, "A", 4)
	testStatement(t, vars, "A+1", 5)
	testStatement(t, vars, "A+A", 8)
	testStatement(t, vars, "B = A*A", 16)
	testStatement(t, vars, "B", 16)
	testStatement(t, vars, "snake_case_69 = B", 16)
	testStatement(t, vars, "snake_case_69 + 1", 17)
}

func TestInvalidSyntax(t *testing.T) {
	assertStatementError(t, "1(")
	assertStatementError(t, ")1")
	assertStatementError(t, "1+")
	assertStatementError(t, "+1")
	assertStatementError(t, "1+1++1")
	assertStatementError(t, "1+1()")
	assertStatementError(t, "1+1 1")
	assertStatementError(t, "1+1+")
	assertStatementError(t, "1+1(")
	assertStatementError(t, "1+1)")
}

func TestFuzzyInput(t *testing.T) {
	seed := rand.Int63()
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
			if err != nil {
				if strings.HasSuffix(err.Error(), "division by 0") {
					continue
				}
				if strings.HasSuffix(err.Error(), "= NaN") {
					continue
				}
				t.Errorf("error: input = %q: %v", input, err)
				return
			}
		}
	})
}

func createRandomGoodInput(random *rand.Rand) []byte {
	n := random.Int()%100 + 1
	digits := []byte("0123456789")
	operators := []string{"+", "-", "*", "/", "//", "%", "v", "**"}
	functions := []string{"asin", "acos", "atan", "sin", "cos", "tan"} // supersets at the end
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

		if random.Intn(10) == 0 && input[len(input)-1] != 'v' {
			fnIdx := random.Intn(len(functions))
			input = append(input, functions[fnIdx]...)
			if random.Intn(2) == 0 {
				input = append(input, ' ')
			} else {
				input = append(input, '(')
			}
		}
	}

	if input[len(input)-1] == '(' || input[len(input)-1] == ' ' {
		input = input[:len(input)-1]
	}
	for _, fn := range functions {
		for bytes.HasSuffix(input, []byte(fn)) {
			input = input[:len(input)-len(fn)]
		}
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
	tokens := []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
		".", "+", "-", "*", "/", "//", "%", "v", "**", "(", ")", " ",
		"sin", "cos", "tan", "asin", "acos", "atan",
		"a", "b", "c", "d",
	}
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

func testStatement(t *testing.T, vars map[string]float64, input string, expected float64) {
	t.Helper()
	errMargin := 0.0000000001
	res, _, processed, err := EvalStatement([]byte(input), vars)
	if err != nil {
		t.Errorf("error: input=%q, processed=%q: %v", input, processed, err)
	}
	if res < expected-errMargin || res > expected+errMargin {
		t.Errorf("calculation error: input=%q, processed=%q: expected %.20f, got %.20f", input, processed, expected, res)
	}
}

func assertStatementError(t *testing.T, input string) {
	t.Helper()
	_, _, _, err := EvalStatement([]byte(input), nil)
	if err == nil {
		t.Errorf("error expected: input=%q", input)
	}
}
