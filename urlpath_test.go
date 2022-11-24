package main

import "testing"

func BenchmarkSplitURLPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		splitURLPath("/foo/bar/baz")
	}
}

// func BenchmarkSplitURLPath2(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		splitURLPath2("/foo/bar/baz")
// 	}
// }
