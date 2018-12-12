Crank is a chaincode test environment and debugger

# Commands

disassemble: disassembles the loaded vm [dis disasm d]

      event: sets the ID of the event to be executed (may change the current IP) [ev e]

       exit: pops the stack; if the top of stack was numeric, uses the lowest byte of its value as the OS exit level [q]
If the top of stack did not exist or was not numeric, exits with 255.
       help: prints this help message (help verbose for extended explanation) [?]

       load: loads the file FILE as a chasm binary (.chbin) [l]
File must conform to the chasm binary standard.
       next: executes one opcode at the current IP and prints the status [n]
If the opcode is a function call, this executes the entire function call before stopping.
        pop: pops the top stack item and prints it [o]

       push: pushes one or more values onto the stack [pu p]

Value syntax:
    Number (decimal, hex)
    Timestamp
    Quoted string
    B(hex pairs)
    [ list of values ] (commas or whitespace, must all be one line)
    { list of values }(commas or whitespace, must all be one line)

       quit: ends the chain program [q]
Ctrl-D also works
      reset: resets the VM to the event and stack that were current at the last Run, Trace, Push, Pop, or Event command [k]

        run: runs the currently loaded VM from the current IP [r]

      stack: prints the contents of the stack [k]

      trace: runs the currently loaded VM from the current IP [tr t]

