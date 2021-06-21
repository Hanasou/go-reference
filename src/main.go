package main

import (
	"fmt"
)

func structEmbedding() {
	fmt.Println("Hello")
	newAb := &AB{
		A: &A{
			Afield: 1,
		},
		B: &B{
			Bfield: 2,
		},
	}
	// Notice how we don't have to refer to Afield through A, even though technically Afield is part of A, not AB.
	fmt.Println(newAb.Afield)
	// Should also apply for methods. We don't have to call getBField() methods through B because of embedding
	fmt.Println(newAb.getBField())
}

func main() {
	multipleIncrementing()
}
