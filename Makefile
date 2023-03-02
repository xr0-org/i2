GO = go
BIN = ${CURDIR}/bin
YACC = $(BIN)/goyacc
PARSER_DIR = ${CURDIR}/internal/parser
SOURCES := $(shell find $(CURDIR) -name '*.go')
I2 = $(BIN)/i2

$(I2): $(PARSER_DIR)/y.go $(SOURCES)
	@printf 'GO\t$@\n'
	@$(GO) build -o $@

check: $(PARSER_DIR)/y.go $(PARSER_DIR)/*_test.go
	@$(GO) test -v ./...

$(PARSER_DIR)/y.go: $(PARSER_DIR)/syntax.go.y $(YACC)
	@printf 'YACC\t$@\n'
	@cd $(PARSER_DIR); $(YACC) -o $@ $<

clean-parser:
	@cd $(PARSER_DIR); rm -f y.*

YACC_DIR = third_party/goyacc
$(YACC): $(YACC_DIR)/yacc.go
	@printf 'GO\t$@\n'
	@cd $(YACC_DIR); $(GO) build -o $@

clean: clean-parser
	@rm -rf $(BIN)
