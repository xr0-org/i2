package truth

func Impl(A, B Proposition) Proposition {
	return implication{A, B}
}

func Fllw(A, B Proposition) Proposition {
	return Impl(B, A)
}

func Not(A Proposition) Proposition {
	return Impl(A, Constant(false))
}

func NotFllw(A, B Proposition) Proposition {
	return Not(Impl(B, A))
}

func Or(A, B Proposition) Proposition {
	return Impl(Impl(A, B), B)
}

func And(A, B Proposition) Proposition {
	return NotFllw(NotFllw(A, B), B)
}

func Eqv(A, B Proposition) Proposition {
	return And(Impl(A, B), Impl(B, A))
}
