package truth

type implication struct {
	antecedent, consequent Proposition
}

func (impl implication) free() []Variable {
	vars := []Variable{}
	m := map[Variable]bool{}
	for _, v := range append(
		impl.antecedent.free(),
		impl.consequent.free()...,
	) {
		if m[v] {
			continue
		}
		m[v] = true
		vars = append(vars, v)
	}
	return vars
}

func (impl implication) eval(m state) bool {
	return !impl.antecedent.eval(m) || impl.consequent.eval(m)
}

func nestedReduce(p Proposition) (Proposition, error) {
	if impl, ok := p.(implication); ok {
		if subred, err := impl.innexReduce(); err == nil {
			return subred, nil
		}
	}
	return nil, errAlreadyPrenex
}

// innexReduce attemts to take one of the reduction steps towards
// expressing the Proposition in prenex normal form, or returns an error
// if this is not possible. In particular, it returns errAlreadyPrenex
// to indicate that the Proposition is in prenex normal form.
func (impl implication) innexReduce() (Proposition, error) {
	if red, err := nestedReduce(impl.antecedent); err == nil {
		return implication{red, impl.consequent}, nil
	}
	if λ, ok := impl.antecedent.(lambda); ok {
		// `(λa)C ==> D` --> `(γb)(C* ==> D)`
		b := λ.reductionVariable(impl.consequent)
		return buildLambda(
			// `(γb)`
			λ.q.flip(), b,
			// `(C* ==> D)`
			implication{λ.scope.replace(λ.v, b), impl.consequent},
		), nil
	}
	if red, err := nestedReduce(impl.consequent); err == nil {
		return implication{impl.antecedent, red}, nil
	}
	if λ, ok := impl.consequent.(lambda); ok {
		// `D ==> (λa)C` --> `(λb)(D ==> C*)`
		b := λ.reductionVariable(impl.antecedent)
		return buildLambda(
			// `(λb)`
			λ.q, b,
			// `(D ==> C*)`
			implication{impl.antecedent, λ.scope.replace(λ.v, b)},
		), nil
	}
	return nil, errAlreadyPrenex
}

func (impl implication) replace(a, b Variable) Proposition {
	return implication{
		impl.antecedent.replace(a, b),
		impl.consequent.replace(a, b),
	}
}

func (impl implication) needsBrackets(op operator) bool {
	if r := resolveImplOp(impl); r.operator.precedence() <= op.precedence() {
		return true
	}
	return false
}

func (impl implication) equals(p Proposition) bool {
	impl2, ok := p.(implication)
	if !ok {
		return false
	}
	return impl.antecedent.equals(impl2.antecedent) &&
		impl.consequent.equals(impl2.consequent)
}

func (impl implication) String() string {
	return formatImplOp(impl)
}
