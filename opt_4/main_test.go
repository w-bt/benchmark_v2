package main

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// go test -v -bench=. -benchtime=30000x -cpuprofile=prof.cpu -memprofile=prof.mem -o bench.test | tee prof.txt
// use -gcflags "-m -l" to do escape analysis
// go tool pprof bench.test prof.cpu
// go tool pprof bench.test prof.mem
// go tool pprof -http=":5001" bench.test prof.cpu
// go tool pprof -http=":5002" bench.test prof.mem
// benchstat prof.txt ../opt_3/prof.txt
func BenchmarkHandleProduct(b *testing.B) {
	b.ReportAllocs()
	r, _ := http.ReadRequest(bufio.NewReader(strings.NewReader("GET /product?code=ZZ99 HTTP/1.0\r\n\r\n")))
	rw := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		handleProduct(rw, r)
		rw.Body.Reset()
	}
}
