package truth

import (
	"testing"
)

func TestElementary(t *testing.T) {
	p, q := Variable("p"), Variable("q")
	// (p ==> q) === !p || q
	impl := Eqv(Impl(p, q), Or(Not(p), q))
	b, err := Decide(impl)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatalf("%s failed", impl)
	}
	r := Variable("r")
	// (p && !r ==> !q) ==> (p && q ==> r)
	impl = Impl(
		// p && !r ==> !q
		Impl(And(p, Not(r)), Not(q)),
		// p && q ==> r
		Impl(And(p, q), r),
	)
	b, err = Decide(impl)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatalf("%s failed", impl)
	}
}

func TestAdvanced(t *testing.T) {
	x, y, z := Variable("x"), Variable("y"), Variable("z")
	/* (∀z) ( !!(∀y)F(y, z) ==> !!(∃x)G(x, y, z) ) */
	Universal("z", Impl(
		// !!(∀y)F(y, z)
		Not(Not(Universal("y",
			Func("F", y, z),
		))),
		// !!(∃x)G(x, y, z)
		Not(Not(Existential("x",
			Func("G", x, y, z),
		))),
	))
}

func TestAdvanced2(t *testing.T) {
	x, p := Variable("x"), Variable("p")
	// p && (∀x)F(x) === (∀x)(p && F(x))
	p0, p1 :=
		// p && (∀x)F(x)
		And(p, Universal("x", Func("F", x))),
		// (∀x)(p && F(x))
		Universal("x", And(p, Func("F", x)))
	Eqv(p0, p1)
}
