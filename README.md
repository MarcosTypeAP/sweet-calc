# Sweet Calc

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
Sweet Calculator REPL
> 1+1
= 2

> bar=24;45+bar
= bar = 24
= 24

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
- `v` Root (eg: `sqrt(9)` = `2v9`)

### Variables

Variable names must start with a letter and can only contain alphanumeric characters and `_`
```bash
> foo = 2+2
= 4

> foo = foo+foo
= 8

> 2+foo
= 10

> foo*foo
= 64

> snake_case_69 = foo
= 8
```
### Syntax sugar
```python
# Space significant
> 1+1 *2
= (1+1)*2
= 4

> 1+1* 2* 4
= 1+1*(2*(4))
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
```

## Installation
```bash
$ git clone git@github.com:MarcosTypeAP/sweet-calc.git --depth=1
$ cd sweetcalc

$ make build # Must have Go installed

$ ./build/c '1+1'
= 2
```
