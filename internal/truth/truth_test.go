package truth

import (
	"testing"
)

func TestImpl(t *testing.T) {
	p, q := variable("p"), variable("q")
	impl := Eqv(Impl(p, q), Or(Not(p), q))
	b, err := decide(impl)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatalf("%s failed", impl)
	}
	impl = Eqv(Impl(Not(p), constant(false)), p)
	b, err = decide(impl)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatalf("%s failed", impl)
	}
}
