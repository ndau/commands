# ndauchasm README

This is the README for the ndauchasm extension for VS Code.

## Features

This extension supports:

* Simple syntax coloring of .chasm files (it also works for the mini-assembler)
* Keyword snippets for opcodes more complex than a single instruction

## Requirements

Copy the entire extension directory to your vscode extensions area. From the cmd
folder:

`cp -R ndauchasm/ ~/.vscode/extensions/oneiro.ndauchasm-0.1.0`

## Known Issues

It doesn't format chasm files because that requires writing JS code. However,
there is also a chfmt command you can use in a parallel folder to this one.

## Release Notes

### 0.1.0.0

Initial release.

