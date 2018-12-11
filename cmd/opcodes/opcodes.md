
# Opcodes for Chaincode

Generated automatically by "make generate"; DO NOT EDIT.

## Implemented and Enabled Opcodes

Value|Opcode|Meaning|Stack before|Instr.|Stack after
----|----|----|----|----|----
0x00|Nop|No-op - has no effect.||nop|
0x01|Drop|Discards the value on top of the stack.|A B|drop|A
0x02|Drop2|Discards the top two values.|A B C|drop2|A
0x05|Dup|Duplicates the top of the stack.|A B|dup|A B B
0x06|Dup2|Duplicates the top two items.|A B C|dup2|A B C B C
0x09|Swap|Exchanges the top two items on the stack.|A B C|swap|A C B
0x0c|Over|Duplicates the second item on the stack to the top of the stack.|A B|over|A B A
0x0d|Pick|The item back in the stack by the specified offset is copied to the top.|A B C D|pick 2|A B C D B
0x0e|Roll|The item back in the stack by the specified offset is moved to the top.|A B C D|roll 2|A C D B
0x0f|Tuck|The top of the stack is dropped N entries back into the stack after removing it from the top.|A B C D|tuck 2|A D B C
0x10|Ret|Terminates the function or handler; the top value on the stack (if there is one) are the return values.||ret|
0x11|Fail|Terminates the function or handler and indicates an error.||fail|
0x1a|One|Pushes 1 onto the stack.||one, true|1
0x1b|Neg1 (True)|Pushes -1 onto the stack.||neg1|-1
0x1c|MaxNum|Pushes the largest possible numeric value onto the stack.||maxnum|9223372036854775807
0x1d|MinNum|Pushes the most negative possible numeric value onto the stack.||minnum|-9223372036854775808
0x20|Zero (False)|Pushes 0 onto the stack.||zero|0
0x21|Push1|Evaluates the next byte as a signed little-endian numeric value and pushes it onto the stack.||push1|A
0x22|Push2|Evaluates the next 2 bytes as a signed little-endian numeric value and pushes it onto the stack.||push2|A
0x23|Push3|Evaluates the next 3 bytes as a signed little-endian numeric value and pushes it onto the stack.||push3|A
0x24|Push4|Evaluates the next 4 bytes as a signed little-endian numeric value and pushes it onto the stack.||push4|A
0x25|Push5|Evaluates the next 5 bytes as a signed little-endian numeric value and pushes it onto the stack.||push5|A
0x26|Push6|Evaluates the next 6 bytes as a signed little-endian numeric value and pushes it onto the stack.||push6|A
0x27|Push7|Evaluates the next 7 bytes as a signed little-endian numeric value and pushes it onto the stack.||push7|A
0x28|Push8|Evaluates the next 8 bytes as a signed little-endian numeric value and pushes it onto the stack.||push8|A
0x2a|PushB|Pushes the specified number of following bytes onto the stack as a Bytes object.||pushb 3 0x41 0x42 0x43|"ABC"
0x2b|PushT|Concatenates the next 8 bytes and pushes them onto the stack as a timestamp.||pusht|timestamp A
0x2c|Now|Pushes the current timestamp onto the stack.||now|(current time as timestamp)
0x2d|PushA|Evaluates a to make sure it is formatted as a valid ndau-style address; if so, pushes it onto the stack as a Bytes object. If not, error.||pusha nda234...4b3|nda234...4b3
0x2e|Rand|Pushes a 64-bit random number onto the stack. Note that 'random' may have special meaning depending on context; in particular, repeated uses of this opcode may (and most likely will) return the same value within a given runtime scenario.||rand|
0x2f|PushL|Pushes an empty list onto the stack.||pushl|[]
0x40|Add|Adds the top two numeric values on the stack and puts their sum on top of the stack. attempting to add non-numeric values is an error.|A B|add|A+B
0x41|Sub|Subtracts the top numeric value on the stack from the second and puts the difference on top of the stack. attempting to subtract non-numeric values is an error.|A B|sub|A-B
0x42|Mul|Multiplies the top two numeric values on the stack and puts their product on top of the stack. attempting to multiply non-numeric values is an error.|A B|mul|A*B
0x43|Div|Divides the second numeric value on the stack by the top and puts the integer quotient on top of the stack. attempting to divide non-numeric values is an error, as is dividing by zero.|A B|div|int(A/B)
0x44|Mod|If the stack has y on top and x in the second position, Mod returns the integer remainder of x/y according to the method that both JavaScript and Go use, which is that it calculates such that q = x/y with the result truncated to zero, where m = x - y*q. The magnitude of the result is less than y and its sign agrees with that of x. Attempting to calculate the mod of non-numeric values is an error. It is also an error if y is zero.|A B|mod|A % B
0x45|DivMod|Divides the second numeric value on the stack by the top and puts the integer quotient on top of the stack and the integer remainder in the second item on the stack, such that q = x/y with the result truncated to zero, where m = x - y*q. Attempting to use non-numeric values is an error, as is dividing by zero.|A B|divmod|A%B int(A/B)
0x46|MulDiv|Multiplies the third numeric item on the stack by the fraction created by dividing the second numeric item by the top; guaranteed not to overflow as long as the fraction is less than 1. An overflow is an error.|A B C|muldiv|int(A*(B/C))
0x48|Not|Evaluates the truthiness of the value on top of the stack, and replaces it with True if the result was False, and with False if the result was True.|5 6 7|not|5 6 0
0x49|Neg|The sign of the number on top of the stack is negated.|A|neg|-A
0x4a|Inc|Adds 1 to the number on top of the stack, which must be a Number.|A|inc|A+1
0x4b|Dec|Subtracts 1 from the number on top of the stack, which must be a Number.|A|dec|A-1
0x50|Index|Selects a zero-indexed element (the index is the top of the stack) from a list reference which is the second item on the stack (both are discarded) and leaves it on top of the stack. Error if index is out of bounds or a list is not the second item.|[X Y Z] 2|index|Z
0x51|Len|Returns the length of a list.|[X Y Z]|len|3
0x52|Append|Creates a new list, appending the new value to it.|[X Y] Z|append|[X Y Z]
0x53|Extend|Generates a new list by concatenating two other lists.|[X Y] [Z]|extend|[X Y Z]
0x54|Slice|Expects a list and two indices on top of the stack. Creates a new list containing the designated subset of the elements in the original slice.|[X Y Z] 1 3|slice|[Y Z]
0x60|Field|Retrieves a field at index f from a struct on top of the stack (which it pops); fails if there is no field at that index or if the top of stack was not a struct.|X|field f|X.f
0x61|IsField|Checks if a field at index f exists in the struct at the top of the stack (which is popped); leaves True if so, False if not. If top was not a struct, fails.|X|isfield f|True if X.f exists
0x70|FieldL|Makes a new list by retrieving a given field from all of the structs in a list.|[X Y Z]|fieldl f|[X.f Y.f Z.f]
0x80|Def|Defines function block n, where n is a number larger than any previously defined function in this script. When the function is called, m values will be copied (not popped) from the caller's stack to a new stack for the use of this function. Functions can only be called by handlers or other functions. Every function must be terminated by enddef, and function definitions may not be nested.||def n m|
0x81|Call|Calls the function block n, provided that its ID is greater than the index of the function block currently executing (recursion is not permitted). The function runs with a new stack which is initialized with the top n values of the current stack (which are copied, NOT popped). Upon return, the top value on the function's stack is pushed onto the caller's stack.||call n|
0x82|Deco|Decorates a list of structs (on top of the stack, which it pops) by applying the function block n to each member of the struct, copying m stack entries (where m is defined by the function) to the function block's stack, then copying the struct itself; on return, that struct's field f is set to the top value of the function's stack. The resulting new list is pushed onto the stack.||deco n f|
0x88|EndDef|Ends a function definition; always required.||enddef|
0x89|IfZ|If the top stack item is zero, executes subsequent code. The top stack item is discarded.||ifz|
0x8a|IfNZ|If the top stack item is nonzero, executes subsequent code. The top stack item is discarded.||ifnz|
0x8e|Else|If the code immediately following an if was not executed, this code (up to end) will be; otherwise it will be skipped.||else|
0x8f|EndIf|Terminates a conditional block; if this opcode is missing for any block, the program is invalid.||endif|
0x90|Sum|Given a list of numbers, sums all the values in the list.|[2 12 4]|sum|18
0x91|Avg|Given a list of numbers, averages all the values in the list. The result will always be Floor(average).|[2 12 4]|avg|6
0x92|Max|Given a list of numbers, finds the maximum value.|[2 12 4]|max|12
0x93|Min|Given a list of numbers, finds the minimum value.|[2 12 4]|min|2
0x94|Choice|Selects an item at random from a list and leaves it on the stack as a replacement for the list.|[X Y Z]|choice|
0x95|WChoice|Selects an item from a list of structs weighted by the given field index, which must be numeric.|[X Y Z] f|wchoice f|
0x96|Sort|Sorts a list of structs by a given field.|[X Y Z] f|sort f|The list sorted by field f
0x97|Lookup|Selects an item from a list of structs by applying the function block n to each item in order, copying m stack entries to the function block's stack (where m is defined by the function), then copying the struct itself; returns the index of the first item in the list where the result is a nonzero number; throws an error if no item returns a nonzero number.|[X Y Z]|lookup n|i
0xa0|Handler|Begins the definition of a handler, which is ended with enddef. The following byte defines a count of the number of handler IDs that follow from 1-255; all of the specified events will be sent to this handler. If the count byte is 0, no handler IDs are specified; this defines the default handler which will receive all events not sent to another handler.||handler 1 EVENT_FOOBAR|
0xb0|Or|Does a bitwise OR of the top two values on the stack (which must both be numeric) and puts the result on top of the stack. Attempting to operate on non-numeric values is an error.|0x55 0x0F|or|0x5F
0xb1|And|Does a bitwise AND of the top two values on the stack (which must both be numeric) and puts the result on top of the stack. Attempting to operate on non-numeric values is an error.|0x55 0x0F|and|0x05
0xb2|Xor|Does a bitwise exclusive OR (XOR) of the top two values on the stack (which must both be numeric) and puts the result on top of the stack. Attempting to operate on non-numeric values is an error.|0x55 0x0F|xor|0x5A
0xbc|Count1s|Returns the number of 1 bits in the top value on the stack (which must be numeric) and puts the result on top of the stack. Attempting to operate on a non-numeric value is an error.|0x55|count1s|4
0xbf|BNot|Does a bitwise NOT (1's complement) of the top value on the stack (which must be numeric) and puts the result on top of the stack. Attempting to operate on a non-numeric value is an error.|5|bnot|-6
0xc0|Lt|Compares (and discards) the two top stack elements. If the types are different, fails execution. If the types are the same, compares the values, and leaves TRUE when the second item is strictly less than the top item according to the comparison rules.|A B|lt|FALSE
0xc1|Lte|Compares (and discards) the two top stack elements. If the types are different, fails execution. If the types are the same, compares the values, and leaves TRUE when the second item is less than or equal to the top item according to the comparison rules.|A B|lte|FALSE
0xc2|Eq|Compares (and discards) the two top stack elements. If the types are different, fails execution. Otherwise, if they are equal in both type and value, leaves TRUE (1) on top of the stack, otherwise leaves FALSE (0) on top of the stack.|A B|eq|FALSE
0xc3|Gte|Compares (and discards) the two top stack elements. If the types are different, fails execution. If the types are the same, compares the values, and leaves TRUE when the second item is greater than or equal to the top item according to the comparison rules.|A B|gte|TRUE
0xc4|Gt|Compares (and discards) the two top stack elements. If the types are different, fails execution. If the types are the same, compares the values, and leaves TRUE when the second item is strictly greater than the top item according to the comparison rules.|A B|gt|TRUE
# Disabled Opcodes

Value|Opcode|Meaning|Stack before|Instr.|Stack after
----|----|----|----|----|----
||There are no disabled opcodes at the moment.||
