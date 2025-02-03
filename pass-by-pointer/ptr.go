package main

type BigStruct struct{ buf [1 << 16]byte }

var obj BigStruct = BigStruct{}

func fooPBV(obj BigStruct)  {}
func fooPBP(obj *BigStruct) {}

func foo() {
	fooPBP(&obj)
	fooPBV(obj)
}

//go test -bench=. -count=1 -gcflags=-N

//Assembly
//go tool compile -N -S -l ptr.go
