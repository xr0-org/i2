package parser

import (
	"fmt"
	"os"

	"git.sr.ht/~lbnz/i2/internal/symbol"
	"git.sr.ht/~lbnz/i2/internal/truth"
)

func Verify(input string) {
	sigma = symbol.Table{"1": symbol.Any}
	if ret := yyParse(&lexer{[]rune(string(input)), 0}); ret != 0 {
		fmt.Fprintf(os.Stderr, "exit code: %d", ret)
	}
}

func sound(prf symbol.RelationChain, tbl symbol.Table) error {
	for _, expr := range prf {
		fmt.Printf("\t%s\n", expr)
		aExpr, err := expr.Analyse(tbl)
		if err != nil {
			return fmt.Errorf("analysis error: %s", err)
		}
		outcome, err := truth.Decide(aExpr.P)
		if err != nil {
			return fmt.Errorf("decision error: %s", err)
		}
		if !outcome {
			return fmt.Errorf("contradiction")
		}
	}
	return nil
}

func getProofProp(A, B truth.Proposition, op symbol.Operator) truth.Proposition {
	switch op {
	case symbol.Eqv, symbol.Impl, symbol.Fllw:
		return op.SimpleTruthOp()(A, B)
	default:
		panic(fmt.Sprintf("invalid op %s", op))
	}
}

func examineProof(assertion symbol.Expr, prf symbol.RelationChain,
	provenLabels []string, tbl symbol.Table) error {
	fmt.Printf("proof:\n")
	// confirm links are valid
	if err := sound(prf, tbl); err != nil {
		return err
	}
	// confirm first and last term joined by appropriate connective imply
	// the asserted proposition
	op, err := prf.GCFOperator()
	if err != nil {
		return err
	}
	provenTbl := symbol.Table{}
	for k, v := range tbl {
		provenTbl[k] = v
	}
	for _, v := range provenLabels {
		provenTbl[v] = symbol.LocalProof{symbol.ConstantExpr(true)}
	}
	first, err := prf[0].E1.Analyse(provenTbl)
	if err != nil {
		return err
	}
	second, err := prf[len(prf)-1].E2.Analyse(provenTbl)
	if err != nil {
		return err
	}
	proofProp := getProofProp(first.P, second.P, op)
	assertionP, err := assertion.Analyse(tbl)
	if err != nil {
		return err
	}
	qed := truth.Impl(proofProp, assertionP.P)
	outcome, err := truth.Decide(qed)
	if err != nil {
		fmt.Println("first", prf[0].E1)
		fmt.Println("prf", prf)
		fmt.Println("assertion", assertionP.P)
		fmt.Println("proof", proofProp)
		fmt.Println("qed was", qed)
		return fmt.Errorf("qed burden failure: %s", err)
	}
	if !outcome {
		return fmt.Errorf("contradiction")
	}
	fmt.Println("qed")
	return nil
}
