# Benchmark and Optimizing Code

Sharing Session Benchmark and Optimizing Code on 05 March 2021 - Groceries Engineering Team

### Requirements

  - Golang 1.11 or newer (I am using 1.14)
  - Benchcmp (go get golang.org/x/tools/cmd/benchcmp)
  - Any editor
  
# Test Case

```golang
package main

import (
	"log"
	"net/http"
	"regexp"
)

var products map[string]*Product

func init() {
	GenerateProduct()
}

func main() {
	log.Printf("Starting on port 1234")
	http.HandleFunc("/product", handleProduct)
	log.Fatal(http.ListenAndServe("127.0.0.1:1234", nil))
}

func handleProduct(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if match, _ := regexp.MatchString(`^[A-Z]{2}[0-9]{2}$`, code); !match {
		http.Error(w, "code is invalid", http.StatusBadRequest)
		return
	}

	result := findProduct(products, code)

	if result.Code == "" {
		http.Error(w, "data not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<font size="10">Product Code : ` + result.Code + ` Name :` + result.Name + `</font>`))
}

func findProduct(Products map[string]*Product, code string) *Product {
	for _, item := range Products {
		if code == (*item).Code {
			return item
		}
	}

	return &Product{}
}
```
### Data

List of product with Code and Name. See [data.go](https://github.com/w-bt/benchmark_v2/blob/master/opt_1/data.go). You may assume this data is comming from database or http call or something else.

### Run It
```sh
$ go build && ./benchmark
```
Open browser and hit `http://localhost:1234/product?code={code}`.

### Test
```sh
$ go test -cover -race -v
?   	github.com/w-bt/benchmark_v2	[no test files]
```

### Add Unit Test and Retest
```golang
package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var tempProducts map[string]*Product

// go test -cover -race -v
func TestMain(m *testing.M) {
	GenerateProduct()

	code := m.Run()

	os.Exit(code)
}

func TestHandleProduct(t *testing.T) {
	testCases := []struct {
		Name            string
		Query           string
		PreExecution    func()
		PostExecution   func()
		SubStringOutput string
	}{
		{
			Name:            "successfully get product code",
			Query:           "AA11",
			PreExecution:    func() {},
			PostExecution:   func() {},
			SubStringOutput: "Product Code",
		},
		{
			Name:            "when regex does not match",
			Query:           "error",
			PreExecution:    func() {},
			PostExecution:   func() {},
			SubStringOutput: "code is invalid",
		},
		{
			Name:  "when data not found",
			Query: "AB11",
			PreExecution: func() {
				tempProducts = products
				products = make(map[string]*Product)
			},
			PostExecution: func() {
				products = tempProducts
			},
			SubStringOutput: "data not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.PreExecution()

			r, _ := http.NewRequest("GET", fmt.Sprintf("/product?code=%s", tc.Query), nil)
			w := httptest.NewRecorder()
			handleProduct(w, r)

			tc.PostExecution()

			require.Equal(t, true, strings.Contains(w.Body.String(), tc.SubStringOutput))
		})
	}
}
```
```sh
$ go test -cover -race -v
=== RUN   TestHandleProduct
--- PASS: TestHandleProduct (0.00s)
PASS
coverage: 73.3% of statements
ok  	github.com/w-bt/benchmark_v2	2.048s
```

### GO Test Benchmark
```golang
func BenchmarkHandleProduct(b *testing.B) {
	b.ReportAllocs()
	r, _ := http.ReadRequest(bufio.NewReader(strings.NewReader("GET /product?code=ZZ99 HTTP/1.0\r\n\r\n")))
	rw := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		handleProduct(rw, r)
		rw.Body.Reset()
	}
}
```
```sh
$ go test -v -bench=. -benchtime=30000x -cpuprofile=prof.cpu -memprofile=prof.mem -o bench.test | tee prof.txt
goos: darwin
goarch: amd64
pkg: github.com/w-bt/benchmark_v2/opt_1
BenchmarkHandleProduct
BenchmarkHandleProduct-12          30000           1058075 ns/op            3354 B/op         49 allocs/op
PASS
ok      github.com/w-bt/benchmark_v2/opt_1      32.077s
```
This command produces 4 new files:
  - binary test (bench.test)
  - cpu profile (prof.cpu)
  - memory profile (prof.mem)
  - benchmark result (prof.txt)

Based on the data above, benchmarking is done by 30000x. Detailed result:
  - total execution: 30000 times
  - duration each operation: 1058075 ns/op
  - each iteration costs: 3354 bytes with 49 heap allocations

### CPU Profiling
```sh
$ go tool pprof bench.test prof.cpu
File: bench.test
Type: cpu
Time: Mar 4, 2021 at 11:36am (WIB)
Duration: 31.67s, Total samples = 30.06s (94.92%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 28.35s, 95.26% of 29.76s total
Dropped 104 nodes (cum <= 0.15s)
Showing top 10 nodes out of 34
      flat  flat%   sum%        cum   cum%
    11.95s 40.15% 40.15%     11.95s 40.15%  memeqbody
     6.81s 22.88% 63.04%      7.79s 26.18%  runtime.mapiternext
     6.09s 20.46% 83.50%     26.40s 88.71%  github.com/w-bt/benchmark_v2/opt_1.findProduct
     1.11s  3.73% 87.23%      1.11s  3.73%  aeshashbody
     0.66s  2.22% 89.45%      0.66s  2.22%  runtime.nanotime1
     0.55s  1.85% 91.30%      0.55s  1.85%  runtime.memequal
     0.47s  1.58% 92.88%      0.47s  1.58%  runtime.madvise
     0.30s  1.01% 93.88%      0.30s  1.01%  runtime.add
     0.22s  0.74% 94.62%      0.22s  0.74%  runtime.(*maptype).indirectkey (inline)
     0.19s  0.64% 95.26%      0.19s  0.64%  runtime.(*bmap).overflow (inline)
(pprof) top5
Showing nodes accounting for 26620ms, 89.45% of 29760ms total
Dropped 104 nodes (cum <= 148.80ms)
Showing top 5 nodes out of 34
      flat  flat%   sum%        cum   cum%
   11950ms 40.15% 40.15%    11950ms 40.15%  memeqbody
    6810ms 22.88% 63.04%     7790ms 26.18%  runtime.mapiternext
    6090ms 20.46% 83.50%    26400ms 88.71%  github.com/w-bt/benchmark_v2/opt_1.findProduct
    1110ms  3.73% 87.23%     1110ms  3.73%  aeshashbody
     660ms  2.22% 89.45%      660ms  2.22%  runtime.nanotime1
(pprof) top --cum
Showing nodes accounting for 24.85s, 83.50% of 29.76s total
Dropped 104 nodes (cum <= 0.15s)
Showing top 10 nodes out of 34
      flat  flat%   sum%        cum   cum%
         0     0%     0%     28.24s 94.89%  github.com/w-bt/benchmark_v2/opt_1.BenchmarkHandleProduct
         0     0%     0%     28.24s 94.89%  github.com/w-bt/benchmark_v2/opt_1.handleProduct
         0     0%     0%     28.24s 94.89%  testing.(*B).runN
         0     0%     0%     28.23s 94.86%  testing.(*B).launch
     6.09s 20.46% 20.46%     26.40s 88.71%  github.com/w-bt/benchmark_v2/opt_1.findProduct
    11.95s 40.15% 60.62%     11.95s 40.15%  memeqbody
     6.81s 22.88% 83.50%      7.79s 26.18%  runtime.mapiternext
         0     0% 83.50%      1.30s  4.37%  runtime.mstart
         0     0% 83.50%      1.12s  3.76%  net/http.Header.Set (inline)
         0     0% 83.50%      1.12s  3.76%  net/textproto.MIMEHeader.Set (inline)
(pprof)
```
Detailed informations:
  - flat: the duration of direct operation inside the function (function call doesn't count as flat)
  - flat%: flat percentage `(flat/total flats)*100`
  - sum%: sum of current flat% and previous flat%, sum% could help you to identify the big rocks quickly
  - cum: cumulative duration for the function, it is the value of the location plus all its descendants.
  - cum%: cumulative percentage `(cum/total cum)*100`

Flat and Sum

Assuming there is a function foo, which is composed of 2 functions and a direct operation.

```
func foo(){
    a()                                 step1
    direct operation                    step2
    b()                                 step3
}
```
In example when we hit foo, it takes 4 seconds, and the time distribution are following.

```
func foo(){
    a()                                 // step1 takes 1s
    direct operation                    // step2 takes 2s
    b()                                 // step3 takes 1s
}
```

flat would be the time spent on step2.
cum would be the total execution time of foo, which contains sub-function call and direct operations. (cum = step1 + step2 + step3)

To see detail duration time for each line, execute this code:
```sh
(pprof) list handleProduct
Total: 29.76s
ROUTINE ======================== github.com/w-bt/benchmark_v2/opt_1.handleProduct in /Users/pt.gojekindonesia/Documents/code/go/src/github.com/w-bt/benchmark_v2/opt_1/main.go
         0     28.24s (flat, cum) 94.89% of Total
         .          .     17:   http.HandleFunc("/product", handleProduct)
         .          .     18:   log.Fatal(http.ListenAndServe("127.0.0.1:1234", nil))
         .          .     19:}
         .          .     20:
         .          .     21:func handleProduct(w http.ResponseWriter, r *http.Request) {
         .       10ms     22:   code := r.FormValue("code")
         .      670ms     23:   if match, _ := regexp.MatchString(`^[A-Z]{2}[0-9]{2}$`, code); !match {
         .          .     24:           http.Error(w, "code is invalid", http.StatusBadRequest)
         .          .     25:           return
         .          .     26:   }
         .          .     27:
         .     26.40s     28:   result := findProduct(products, code)
         .          .     29:
         .          .     30:   if result.Code == "" {
         .          .     31:           http.Error(w, "data not found", http.StatusBadRequest)
         .          .     32:           return
         .          .     33:   }
         .          .     34:
         .      1.13s     35:   w.Header().Set("Content-Type", "text/html; charset=utf-8")
         .       30ms     36:   w.Write([]byte(`<font size="10">Product Code : ` + result.Code + ` Name :` + result.Name + `</font>`))
         .          .     37:}
         .          .     38:
         .          .     39:func findProduct(Products map[string]*Product, code string) *Product {
         .          .     40:   for _, item := range Products {
         .          .     41:           if code == (*item).Code {
(pprof) list findProduct
Total: 29.76s
ROUTINE ======================== github.com/w-bt/benchmark_v2/opt_1.findProduct in /Users/pt.gojekindonesia/Documents/code/go/src/github.com/w-bt/benchmark_v2/opt_1/main.go
     6.09s     26.40s (flat, cum) 88.71% of Total
         .          .     35:   w.Header().Set("Content-Type", "text/html; charset=utf-8")
         .          .     36:   w.Write([]byte(`<font size="10">Product Code : ` + result.Code + ` Name :` + result.Name + `</font>`))
         .          .     37:}
         .          .     38:
         .          .     39:func findProduct(Products map[string]*Product, code string) *Product {
     750ms      8.54s     40:   for _, item := range Products {
     5.34s     17.86s     41:           if code == (*item).Code {
         .          .     42:                   return item
         .          .     43:           }
         .          .     44:   }
         .          .     45:
         .          .     46:   return &Product{}
(pprof)
```

To see in UI form, use `web`

![cpu profile](./opt_1/pprof001.svg)

#### Interpreting the Callgraph

Node Color:
-  large positive cum values are red.
-  cum values close to zero are grey.

Node Font Size:
-  larger font size means larger absolute flat values.
-  smaller font size means smaller absolute flat values.

Edge Weight:
-  thicker edges indicate more resources were used along that path.
-  thinner edges indicate fewer resources were used along that path.

Edge Color:
-  large positive values are red.
-  values close to zero are grey.

Dashed Edges: some locations between the two connected locations were removed.

Solid Edges: one location directly calls the other.

"(inline)" Edge Marker: the call has been inlined into the caller.

### Memory Profiling

Similar with CPU Profiling, execute this command
```sh
go tool pprof bench.test prof.mem
File: bench.test
Type: alloc_space
Time: Mar 4, 2021 at 11:29pm (WIB)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top --cum
Showing nodes accounting for 35.50MB, 34.27% of 103.59MB total
Dropped 3 nodes (cum <= 0.52MB)
Showing top 10 nodes out of 43
      flat  flat%   sum%        cum   cum%
         0     0%     0%    92.52MB 89.31%  github.com/w-bt/benchmark_v2/opt_1.BenchmarkHandleProduct
       2MB  1.93%  1.93%    92.52MB 89.31%  github.com/w-bt/benchmark_v2/opt_1.handleProduct
         0     0%  1.93%    92.52MB 89.31%  testing.(*B).launch
         0     0%  1.93%    92.52MB 89.31%  testing.(*B).runN
         0     0%  1.93%    90.52MB 87.38%  regexp.Compile (inline)
         0     0%  1.93%    90.52MB 87.38%  regexp.MatchString
       4MB  3.86%  5.79%    90.52MB 87.38%  regexp.compile
       2MB  1.93%  7.72%       33MB 31.86%  regexp/syntax.Parse
   27.50MB 26.55% 34.27%    27.50MB 26.55%  regexp/syntax.(*parser).newRegexp (inline)
         0     0% 34.27%    24.01MB 23.17%  regexp.compileOnePass
(pprof) list handleProduct
Total: 103.59MB
ROUTINE ======================== github.com/w-bt/benchmark_v2/opt_1.handleProduct in /Users/pt.gojekindonesia/Documents/code/go/src/github.com/w-bt/benchmark_v2/opt_1/main.go
       2MB    92.52MB (flat, cum) 89.31% of Total
         .          .     18:   log.Fatal(http.ListenAndServe("127.0.0.1:1234", nil))
         .          .     19:}
         .          .     20:
         .          .     21:func handleProduct(w http.ResponseWriter, r *http.Request) {
         .          .     22:   code := r.FormValue("code")
         .    90.52MB     23:   if match, _ := regexp.MatchString(`^[A-Z]{2}[0-9]{2}$`, code); !match {
         .          .     24:           http.Error(w, "code is invalid", http.StatusBadRequest)
         .          .     25:           return
         .          .     26:   }
         .          .     27:
         .          .     28:   result := findProduct(products, code)
         .          .     29:
         .          .     30:   if result.Code == "" {
         .          .     31:           http.Error(w, "data not found", http.StatusBadRequest)
         .          .     32:           return
         .          .     33:   }
         .          .     34:
         .          .     35:   w.Header().Set("Content-Type", "text/html; charset=utf-8")
       2MB        2MB     36:   w.Write([]byte(`<font size="10">Product Code : ` + result.Code + ` Name :` + result.Name + `</font>`))
         .          .     37:}
         .          .     38:
         .          .     39:func findProduct(Products map[string]*Product, code string) *Product {
         .          .     40:   for _, item := range Products {
         .          .     41:           if code == (*item).Code {
(pprof) list MatchString
Total: 103.59MB
ROUTINE ======================== regexp.MatchString in /usr/local/Cellar/go/1.14.2/libexec/src/regexp/regexp.go
         0    90.52MB (flat, cum) 87.38% of Total
         .          .    526:
         .          .    527:// MatchString reports whether the string s
         .          .    528:// contains any match of the regular expression pattern.
         .          .    529:// More complicated queries need to use Compile and the full Regexp interface.
         .          .    530:func MatchString(pattern string, s string) (matched bool, err error) {
         .    90.52MB    531:   re, err := Compile(pattern)
         .          .    532:   if err != nil {
         .          .    533:           return false, err
         .          .    534:   }
         .          .    535:   return re.MatchString(s), nil
         .          .    536:}
(pprof) web handleProduct
```
![mem profile](./opt_1/pprof002.svg)

### Another Way

Use Web Version!!!!

```sh
$ go tool pprof -http=":8081" [binary] [profile]
```

# Optimization

### Update findProduct()

```golang
func findProduct(Products map[string]*Product, code string) Product {
	if v, ok := Products[code]; ok {
		return *v
	}

	return Product{}
}
```
```sh
$ go test -v -bench=. -benchtime=30000x -cpuprofile=prof.cpu -memprofile=prof.mem -o bench.test | tee prof.txt
goos: darwin
goarch: amd64
pkg: github.com/w-bt/benchmark/opt_2
BenchmarkHandleProduct
BenchmarkHandleProduct-12          30000              5166 ns/op            3353 B/op         49 allocs/op
PASS
ok      github.com/w-bt/benchmark/opt_2 0.411s
```
```sh
$ benchcmp prof.txt ../opt_1/prof.txt       
benchmark                     old ns/op     new ns/op     delta
BenchmarkHandleProduct-12     5166          1058075       +20381.51%

benchmark                     old allocs     new allocs     delta
BenchmarkHandleProduct-12     49             49             +0.00%

benchmark                     old bytes     new bytes     delta
BenchmarkHandleProduct-12     3353          3354          +0.03%
```

### Update Regex

Compile regex first
```golang
	codeRegex = regexp.MustCompile(`^[A-Z]{2}[0-9]{2}$`)
	// . . .
	if match := codeRegex.MatchString(code); !match {
		// . . .
	}
```

```sh
$ go test -v -bench=. -benchtime=30000x -cpuprofile=prof.cpu -memprofile=prof.mem -o bench.test | tee prof.txt
goos: darwin
goarch: amd64
pkg: github.com/w-bt/benchmark/opt_3
BenchmarkHandleProduct
BenchmarkHandleProduct-12          30000               483 ns/op             144 B/op          3 allocs/op
PASS
ok      github.com/w-bt/benchmark/opt_3 1.549s
```
```sh
$ benchcmp prof.txt ../opt_2/prof.txt
benchmark                     old ns/op     new ns/op     delta
BenchmarkHandleProduct-12     483           5166          +969.57%

benchmark                     old allocs     new allocs     delta
BenchmarkHandleProduct-12     3              49             +1533.33%

benchmark                     old bytes     new bytes     delta
BenchmarkHandleProduct-12     144           3353          +2228.47%
```
### Update Concate String
```golang
var buf       = new(bytes.Buffer)
// . . .
buf.Reset()
buf.WriteString(`<font size="10">Product Code : `)
buf.WriteString(result.Code)
buf.WriteString(` Name :`)
buf.WriteString(result.Name)
buf.WriteString(`</font>`)
w.Write(buf.Bytes())
```
```sh
$ go test -v -bench=. -benchtime=30000x -cpuprofile=prof.cpu -memprofile=prof.mem -o bench.test | tee prof.txt
goos: darwin
goarch: amd64
pkg: github.com/w-bt/benchmark/opt_4
BenchmarkHandleProduct
BenchmarkHandleProduct-12          30000               270 ns/op              16 B/op          1 allocs/op
PASS
ok      github.com/w-bt/benchmark/opt_4 1.553s
```
```sh
$ benchcmp prof.txt ../opt_3/prof.txt
benchmark                     old ns/op     new ns/op     delta
BenchmarkHandleProduct-12     270           483           +78.89%

benchmark                     old allocs     new allocs     delta
BenchmarkHandleProduct-12     1              3              +200.00%

benchmark                     old bytes     new bytes     delta
BenchmarkHandleProduct-12     16            144           +800.00%
```
