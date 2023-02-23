package truth

import (
	"testing"
)

func TestImpl(t *testing.T) {
	p, q := Variable("p"), Variable("q")
	impl := Eqv(Impl(p, q), Or(Not(p), q))
	b, err := Decide(impl)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatalf("%s failed", impl)
	}
	r := Variable("q")
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
