package main

// A Command acts like a sub-program: it can specify its own argument parser,
// and can succeed or fail, displaying arbitrary messages as a result, without
// exiting the shell.
type Command interface {
	// Name is the name by which this command is called.
	//
	// If more than one space-separated word is present, it is treated as a synonym.
	Name() string

	// Run invokes this command.
	//
	// Much like standard commands on the CLI, the run string is the entire
	// string typed by the user, so it will always begin with the name of
	// the command. Lexing is handled by the shell, but argument parsing is
	// the responsibility of the individual command.
	//
	// Unlike typical CLI commands, commands within ndsh have mutable access
	// to typed global state. They may also return human-readable errors.
	//
	// Long-running commands are encouraged to background themselves by doing
	// most of their work in a separate goroutine and returning immediately.
	// If they choose to do so, they should coordinate by adding to the
	// shell.Running WaitGroup before launching their goroutine and
	// defer shell.Running.Done() immediately after launching it.
	//
	// All commands taking more than trivial time, or which perform network IO,
	// should periodically attempt to read from shell.Stop; if such a read
	// produces a result, the command should shut itself down immediately.
	Run([]string, *Shell) error
}
