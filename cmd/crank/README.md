Crank is a chaincode test environment and debugger.

It has some basic capabilities as a REPL and debugger, but has become mainly about writing automated tests for chaincode.

You can get basic help by running it and typing `help` or `help verbose`.

# Basic usage as a REPL

## start crank and run a simple test

given this chasm program:
```
; Demonstration code for chasm assembler
; This program expects a, b, c,
; and x on the stack and calculates
; a*x*x + b*x + c

handler EVENT_DEFAULT {
                                    ; ABCX
    roll 3                          ; BCXA
    pick 1                          ; BCXAX
    dup                             ; BCXAXX
    mul                             ; BCXAR
    mul                             ; BCXR
    roll 3                          ; CXRB
    roll 2                          ; CRBX
    mul                             ; CRS
    add                             ; CR
    add                             ; R
}
```

After assembly, we can run it as follows:

```
$ cmd/crank/crank -binary cmd/chasm/examples/quadratic.chbin
02:  0e 03                     Roll    03            STK: Empty

  1 crank> push 1 2 3
02:  0e 03                     Roll    03            STK: 3, 2, 1

  2 crank> push 5
02:  0e 03                     Roll    03            STK: 5, 3, 2, 1

  3 crank> run
END                                      STK: 38

  4 crank>
```

# Commands

## disassemble
(also `dis`, `disasm`, or `d`)

Disassembles the entire loaded vm.

## event
(also `ev` and `e`)

Sets the ID of the event to be executed (will set the current IP to the start of the identified handler). The event name can be a number or the constant name starting with EVENT_ (see the `const` command).

## exit

Pops the stack; if the top of stack was numeric, uses the lowest byte of its value as the OS exit level. If the top of stack did not exist or was not numeric, exits with 255.

## help
(also `?`)

Prints a help message (help verbose for extended explanation)

## load
(also `l`)
Loads the file FILE as a chasm binary (.chbin) file. File must conform to the chasm binary standard.
The file is opened relative to the current directory, but if the load command is in a script, if the load fails, it also attempts to open the file relative to the script's directory.

## next
(also `n`)

Executes one opcode at the current IP and prints the status. If the opcode is a function call, this executes the entire function call before stopping. (It basically does a step over rather than a step in. Someday we may allow both.)

## pop
(also `o`)

Pops the top stack item and prints it.

## push
(also `pu`, `p`)

Pushes one or more values onto the stack.

Value syntax:
* Number
    * Decimal - 193
    * Hex - 0xD1CE
    * napu - np500 = 500 napu
    * ndau - nd2.5 = 250000000 napu
* Timestamp - 2018-12-17T10:24:33Z
* Quoted string -- single, double, or triplequotes of either kind, no escape characters
* B(hex pairs) -- B(12345678abcdef)
* [ list of values ] (commas or whitespace)
* { collection of index:value pairs }(commas or whitespace)
* account -- this single word generates a random account object and pushes it

Push is a single-line command, but can parse values of arbitrary complexity.

## quit
(also `q`)

Ends the crank program without inspecting the stack.
Ctrl-D (EOF) also works.

If a script hits EOF without a quit instruction, it drops into the REPL. This is also true if a script encounters an unexpected error return from a run command. To end a script and exit to the operating system, use the quit command.

## reset

Resets the VM to the event and stack that were current at the last `run`, `trace`, `push`, `pop`, or `event` command.

## clear
Clears (empties) the stack. Use this in scripts before attempting additional runs.

## run [fail | succeed]
(also `r`)

Runs the currently loaded VM starting at the current IP. If either `fail` or `succeed` is specified, the run is expected to terminate with the given status; if it does not, the crank program will exit with an error code (or drop into the REPL if the `verbose` flag is set).

## stack
(also `k`)

Prints the contents of the stack [k]

## trace
(also `tr`, `t`)

Runs the currently loaded VM from the current IP but in single step mode, disassembling each instruction as it proceeds.

## constants
(also `const`)

Prints a list of the predefined constants. If it is followed by text, only the constants containing that text will be printed.

## expect [value...]

Mainly for use in scripts. If the expected values are not found or an error occurs, exits with a nonzero return code.

## ;
Blank lines are ignored. Comments start with a semicolon and continue to EOL.

# Use as a test script runner

## Script format
Write your scripts like this:

```
; always start with a relative load
load ./ntrd_target_sales.chbin

; validation scripts expect an account, a tx, and a bitmask
; but if our script doesn't inspect them
; we can just push a 0 rather than an account record

clear
push 0
push { TX_QUANTITY: nd100 }
push 0x05
event EVENT_CHANGEVALIDATION
run succeed

clear
push 0
push { TX_QUANTITY: nd1.5 }
push 0x01
event EVENT_CHANGEVALIDATION
run fail

; always end with quit
quit
```

## Notes on scripts

* Most output is suppressed when running scripts, but `run` and `expect` will dump errors that disagree with their parameters and immediately terminate the run, setting the error code.
* The script line number is included in the error code.
* If you use the -verbose (-v) switch on the command line, instead of terminating, a failure will terminate into the repl so you can inspect the state or try again.


## Todo
* Add history command since VM supports history
* Use a more structured disassembly
* Give chasm the ability to create an annotated listing file and use the listing instead of disassembly
