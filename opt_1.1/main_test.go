package main

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var tempProducts map[string]Product

// go test -v -bench=. -benchtime=30000x -cpuprofile=prof.cpu -memprofile=prof.mem -o bench.test | tee prof
// use -gcflags "-m -l" to do escape analysis
// go tool pprof bench.test prof.cpu
// go tool pprof bench.test prof.mem
// go tool pprof -http=":5001" bench.test prof.cpu
// go tool pprof -http=":5002" bench.test prof.mem
// benchcmp prof ../opt_1/prof
func BenchmarkHandleProduct(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r, _ := http.ReadRequest(bufio.NewReader(strings.NewReader("GET /product?code=ZZ99 HTTP/1.0\r\n\r\n")))
		rw := httptest.NewRecorder()
		handleProduct(rw, r)
		rw.Body.Reset()
	}
}
