# Sweet Calc

This is a handy **CLI** calculator for doing quick calculations with **custom operators** and some **syntax sugar**.

## Usage
```bash
# From stdin
$ echo '1+1;2+2' | c
= 2
= 4

# With argument
$ c 1+1
= 2

# Or interactively as REPL
$ c
> 1+1
= 2

> bar=24;45+bar
= bar = 24
= 69

# Up/Down Arrows to navigate history
> bar=24;45+bar

# Delete input with Ctrl+D
> |

# Quit with Ctrl-C
$ |
```

> ⚠️ REPL mode is only available for Unix at the moment.

### Operators

- `+` Addition
- `-` Subtraction
- `*` Multiplication
- `%` Modulo
- `/` Division
- `//` Floor Division
- `**` Power
- `v` Root (eg: `sqrt(9)` = `2v9`)

### Functions

- `sin` Sine
- `cos` Cosine
- `tan` Tangent
- `asin` Arc sine
- `acos` Arc cosine
- `atan` Arc tangent

### Variables

Variable names must start with a letter and can only contain alphanumeric characters and `_`
```bash
> foo = 2+2
= foo = 4

> foo = foo+foo
= foo = 8

> sin 2+foo
= -0.544021

> foo*foo
= 64

> snake_case_69 = foo
= snake_case_69 = 8
```

### Numbers

Floats can start with `.` and the integer part can be spaced with `_`
```bash
> .2
= 0.2

> 0.100001
= 0.100001

# Display precision of 6 decimal places
> 0.1000001
= 0.1

> 2v2
= 1.414214

> 69_420
= 69_420

> 33___33
= 3333

> 1000000.21
= 1_000_000.21
```

### Syntax sugar

```bash
# Space significant
> 1+1 *2
# (1+1)*2
= 4

> 1+1* 2* 4
# 1+1*(2*(4))
= 9

# Multiplication when next to bracket
> 2(1+1)
= 4

> (1+1)(1+1)
= 4

# Optional closing brackets
> 8/(1+1
= 4
```
### Errors
```bash
# Nice errors
$ c '1+1 1+1'

    1+1 1+1
      ^^^
error at position 2:
    preprocessor: token: 3: two consecutive operands without operator
    
$ c '1+1* bar+1'

    1+1*(bar+1)
         ^^^
error at position 5:
    eval tree: undefined variable: "bar"

$ c '2*2+asin 2'

    2*2+asin(2)
        ^^^^
error at position 4:
    eval tree: asin(2) = NaN
```

## Installation
```bash
$ git clone git@github.com:MarcosTypeAP/sweet-calc.git --depth=1
$ cd sweetcalc

# Must have Go installed
$ PREFIX=~/.local/bin make install

$ c '1+1'
= 2
```
