package truth

func Impl(A, B proposition) proposition {
	return implication{A, B}
}

func Fllw(A, B proposition) proposition {
	return Impl(B, A)
}

func Not(A proposition) proposition {
	return Impl(A, constant(false))
}

func NotFllw(A, B proposition) proposition {
	return Not(Impl(B, A))
}

func Or(A, B proposition) proposition {
	return Impl(Impl(A, B), B)
}

func And(A, B proposition) proposition {
	return NotFllw(NotFllw(A, B), B)
}

func Eqv(A, B proposition) proposition {
	return And(Impl(A, B), Impl(B, A))
}
