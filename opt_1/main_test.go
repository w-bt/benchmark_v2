package main

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

//var tempProducts map[string]*Product
//
//// go test -cover -race -v
//func TestMain(m *testing.M) {
//	GenerateProduct()
//
//	code := m.Run()
//
//	os.Exit(code)
//}
//
//func TestHandleProduct(t *testing.T) {
//	testCases := []struct {
//		Name            string
//		Query           string
//		PreExecution    func()
//		PostExecution   func()
//		SubStringOutput string
//	}{
//		{
//			Name:            "successfully get product code",
//			Query:           "AA11",
//			PreExecution:    func() {},
//			PostExecution:   func() {},
//			SubStringOutput: "Product Code",
//		},
//		{
//			Name:            "when regex does not match",
//			Query:           "error",
//			PreExecution:    func() {},
//			PostExecution:   func() {},
//			SubStringOutput: "code is invalid",
//		},
//		{
//			Name:  "when data not found",
//			Query: "AB11",
//			PreExecution: func() {
//				tempProducts = products
//				products = make(map[string]*Product)
//			},
//			PostExecution: func() {
//				products = tempProducts
//			},
//			SubStringOutput: "data not found",
//		},
//	}
//
//	for _, tc := range testCases {
//		t.Run(tc.Name, func(t *testing.T) {
//			tc.PreExecution()
//
//			r, _ := http.NewRequest("GET", fmt.Sprintf("/product?code=%s", tc.Query), nil)
//			w := httptest.NewRecorder()
//			handleProduct(w, r)
//
//			tc.PostExecution()
//
//			require.Equal(t, true, strings.Contains(w.Body.String(), tc.SubStringOutput))
//		})
//	}
//}

// go test -v -bench=. -benchtime=30000x -cpuprofile=prof.cpu -memprofile=prof.mem -o bench.test | tee prof
// use -gcflags "-m -l" to do escape analysis
// go tool pprof bench.test prof.cpu
// go tool pprof bench.test prof.mem
// go tool pprof -http=":5001" bench.test prof.cpu
// go tool pprof -http=":5002" bench.test prof.mem
func BenchmarkHandleProduct(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r, _ := http.ReadRequest(bufio.NewReader(strings.NewReader("GET /product?code=ZZ99 HTTP/1.0\r\n\r\n")))
		rw := httptest.NewRecorder()
		handleProduct(rw, r)
		rw.Body.Reset()
	}
}
