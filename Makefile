###################################
# This Makefile builds several of the chaincode tools.

# Chaincode depends on the opcodes project, which uses a template system to generate
# sources for several things -- the documentation for chaincode as well as much of the
# implementation of both the VM and the assembler. This ensures that the various parts
# don't get out of sync with each other.
###################################

# define a few of the executables we're building
CHASM = cmd/chasm/chasm
CHAIN = cmd/chain/chain
CRANK = cmd/crank/crank
CRANKGEN = cmd/crank/crankgen.py
CHFMT = cmd/chfmt/chfmt
PEGGOFMT = cmd/peggofmt/peggofmt
EXAMPLES = cmd/chasm/examples
OPCODES = cmd/opcodes/opcodes
OPCODESMD = cmd/opcodes/opcodes.md

# And identify the locations of related packages
CHAINCODEPKG = ../chaincode/pkg
VALIDATIONS = ../validation_scripts

###################################
### Some conveniences

.PHONY: generate clean fuzz fuzzmillion benchmarks \
	test examples chaincodeall build chasm crank chfmt \
	opcodes format validations vtests vformat vgen vclean

opcodes: $(OPCODES)

crank: $(CRANK)

chasm: $(CHASM)

chfmt: $(CHFMT)

peggofmt: $(PEGGOFMT)

###################################
### Utilities

default: build

setup:
	hash pigeon
	hash msgp
	hash stringer
	go get $(CHAINCODEPKG)/...

clean:
	rm -f $(OPCODES)
	rm -f $(CHASM)
	rm -f $(CRANK)
	rm -f $(CHFMT)
	rm -f $(PEGGOFMT)
	# generated files
	rm -f cmd/chasm/chasm.go
	rm -f cmd/chfmt/chfmt.go

build: generate opcodes chasm crank chfmt

test: cmd/chasm/chasm.go $(CHAINCODEPKG)/vm/*.go $(CHAINCODEPKG)/chain/*.go chasm
	rm -f /tmp/cover*
	go test $(CHAINCODEPKG)/chain -v --race -timeout 10s -coverprofile=/tmp/coverchain
	go test ./cmd/chasm -v --race -timeout 10s -coverprofile=/tmp/coverchasm
	go test $(CHAINCODEPKG)/vm -v --race -timeout 10s -coverprofile=/tmp/covervm

chaincodeall: clean generate build test fuzz benchmarks format examples

###################################
### Opcodes

$(OPCODESMD): $(OPCODES)
	$(OPCODES) --opcodes $(OPCODESMD)

$(CHAINCODEPKG)/vm/opcodes.go: $(OPCODES)
	$(OPCODES) --defs $(CHAINCODEPKG)/vm/opcodes.go

$(CHAINCODEPKG)/vm/miniasmOpcodes.go: $(OPCODES)
	$(OPCODES) --miniasm $(CHAINCODEPKG)/vm/miniasmOpcodes.go

$(CHAINCODEPKG)/vm/extrabytes.go: $(OPCODES)
	$(OPCODES) --extra $(CHAINCODEPKG)/vm/extrabytes.go

$(CHAINCODEPKG)/vm/enabledopcodes.go: $(OPCODES)
	$(OPCODES) --enabled $(CHAINCODEPKG)/vm/enabledopcodes.go

cmd/chasm/chasm.peggo: $(OPCODES) $(PEGGOFMT)
	$(OPCODES) --pigeon cmd/chasm/chasm.peggo
	$(PEGGOFMT) cmd/chasm/chasm.peggo

# We make two copies of this file, for chasm and for crank
cmd/chasm/predefined.go: $(OPCODES)
	$(OPCODES) --consts cmd/chasm/predefined.go

cmd/crank/predefined.go: $(OPCODES)
	$(OPCODES) --consts cmd/crank/predefined.go

$(OPCODES): cmd/opcodes/*.go
	cd cmd/opcodes && go build

###################################
### The vm itself and its tests

generate: $(OPCODESMD) $(CHAINCODEPKG)/vm/opcodes.go \
		$(CHAINCODEPKG)/vm/miniasmOpcodes.go $(CHAINCODEPKG)/vm/opcode_string.go \
		$(CHAINCODEPKG)/vm/extrabytes.go $(CHAINCODEPKG)/vm/enabledopcodes.go \
		cmd/chasm/chasm.peggo cmd/chasm/predefined.go cmd/crank/predefined.go

$(CHAINCODEPKG)/vm/opcode_string.go: $(CHAINCODEPKG)/vm/opcodes.go
	go generate $(CHAINCODEPKG)/vm

fuzz: test
	FUZZ_RUNS=10000 go test --race -v -timeout 1m $(CHAINCODEPKG)/vm -run "TestFuzz*" -coverprofile=/tmp/coverfuzz

fuzzmillion: test
	FUZZ_RUNS=1000000 go test --race -v -timeout 2h $(CHAINCODEPKG)/vm -run "TestFuzz*" -coverprofile=/tmp/coverfuzz

benchmarks:
	cd $(CHAINCODEPKG)/vm && go test -bench=. -benchmem

###################################
### The chasm assembler

$(CHASM): cmd/chasm/chasm.go $(CHAINCODEPKG)/vm/opcodes.go cmd/chasm/*.go
	go build -o $(CHASM) ./cmd/chasm

cmd/chasm/chasm.go: cmd/chasm/chasm.peggo
	pigeon -o ./cmd/chasm/chasm.go ./cmd/chasm/chasm.peggo

examples: $(CHASM)
	$(CHASM) --output $(EXAMPLES)/quadratic.chbin --comment "Test of quadratic" $(EXAMPLES)/quadratic.chasm
	$(CHASM) --output $(EXAMPLES)/majority.chbin --comment "Test of majority" $(EXAMPLES)/majority.chasm
	$(CHASM) --output $(EXAMPLES)/onePlus1of3.chbin --comment "1+1of3" $(EXAMPLES)/onePlus1of3.chasm
	$(CHASM) --output $(EXAMPLES)/first.chbin --comment "the first key must be set" $(EXAMPLES)/first.chasm
	$(CHASM) --output $(EXAMPLES)/one.chbin --comment "unconditionally return numeric 1" $(EXAMPLES)/one.chasm
	$(CHASM) --output $(EXAMPLES)/zero.chbin --comment "returns numeric 0 in all cases" $(EXAMPLES)/zero.chasm
	$(CHASM) --output $(EXAMPLES)/rfe.chbin --comment "standard RFE rules" $(EXAMPLES)/rfe.chasm

vclean:
	find $(VALIDATIONS) -name "*gen.crank" -print0 | xargs -0 rm

validations: $(CHASM)
	find $(VALIDATIONS) -name "*.chasm" |sed s/\.chasm/.ch/g | xargs -n1 -I{} $(CHASM) --output {}bin {}asm

vgen: $(CRANK) validations vclean
	find $(VALIDATIONS) -name "*.crankgen" -print0 | xargs -0 $(CRANKGEN)

vtests: $(CRANK) vgen
	find $(VALIDATIONS) -name "*.crank" -print0 | xargs -0 -n1 -I{} $(CRANK) -script {}

vformat: $(CHFMT) validations
	find $(VALIDATIONS) -name "*.chasm" -print0 | xargs -0 -n1 -I{} $(CHFMT) -O {}

###################################
### The chfmt formatter

format: $(CHFMT)
	$(CHFMT) -O $(EXAMPLES)/quadratic.chasm
	$(CHFMT) -O $(EXAMPLES)/majority.chasm
	$(CHFMT) -O $(EXAMPLES)/onePlus1of3.chasm
	$(CHFMT) -O $(EXAMPLES)/first.chasm
	$(CHFMT) -O $(EXAMPLES)/one.chasm
	$(CHFMT) -O $(EXAMPLES)/zero.chasm
	$(CHFMT) -O $(EXAMPLES)/rfe.chasm

cmd/chfmt/chfmt.go: cmd/chfmt/chfmt.peggo
	pigeon -o ./cmd/chfmt/chfmt.go ./cmd/chfmt/chfmt.peggo

$(CHFMT): cmd/chfmt/*.go cmd/chfmt/chfmt.go
	go build -o $(CHFMT) ./cmd/chfmt


###################################
### The peggofmt formatter

cmd/peggofmt/peggo.go: cmd/peggofmt/peggo.peggo
	pigeon -o ./cmd/peggofmt/peggo.go ./cmd/peggofmt/peggo.peggo

$(PEGGOFMT): cmd/peggofmt/*.go cmd/peggofmt/peggo.go
	go build -o $(PEGGOFMT) ./cmd/peggofmt


###################################
### The crank debugger/runtime

$(CRANK): cmd/crank/*.go
	go build -o $(CRANK) ./cmd/crank

