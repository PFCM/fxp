// package vgen implements a small dsl for building, comparing and testing
// optimised routines for operating on our 8 bit data types. The routines
// all have some number of inputs and a single output.
// The grammar is:
//
//	 func = (inp "\n")+ ret "\n" [vstmt "\n"]+ stmt "\n"
//	  inp = "input:" " " ident " " type
//	ident = /[a-zA-Z][a-zA-Z0-9]+/
//	 type = ["v"] "s17"
//	  ret = "return:" " " type
//	vstmt = ident " " "=" " " stmt
//	 stmt =
//
// where "+" means "at least one" and " " means any number of spaces or tabs.
package vgen

// Node is a node in the computation graph.
type Node interface{}

// VS17Add adds two vectors of S17s.
type VS17Add struct {
	left, right Node
}
