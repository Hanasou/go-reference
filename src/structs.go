package main

type A struct {
	Afield int
}

type B struct {
	Bfield int
}

// If we want to embed structs in other structs, then we omit its type
type AB struct {
	*A
	*B
}

func (a *A) getAField() int {
	return a.Afield
}

func (b *B) getBField() int {
	return b.Bfield
}
