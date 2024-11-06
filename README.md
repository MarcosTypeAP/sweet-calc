# SweetCalc

This is a handy **CLI** calculator for doing quick calculations with **custom operators** and some **syntax sugar**.

## Usage
```bash
# From stdin
$ echo 1+1 | c
= 2

# With argument
$ c 1+1
= 2

# Or interactively as REPL
$ c
1+1
= 2
45+24
= 69
```
### Operators

- `+` Addition
- `-` Subtraction
- `*` Multiplication
- `%` Modulo
- `/` Division
- `//` Floor Division
- `**` Power
- `v` Root (eg: sqrt(9) = 2v9)

### Syntax sugar
```python
# Space significant
1+1 *2
= (1+1)*2
= 4

1+1* 2* 4
= 1+1*(2*(4))
= 9

# Multiplication when next to bracket
2(1+1)
= 4

(1+1)(1+1)
= 4

# Optional closing brackets
2/(1+1
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
    
 $ c '1+1A0'

    1+1A0
       ^
error at position 3:
    lexer: char 3: unexpected character
```

## Installation
```bash
$ git clone git@github.com:MarcosTypeAP/sweetcalc.git --depth=1
$ cd sweetcalc

$ make build

$ ./build/c '1+1'
= 2
```
