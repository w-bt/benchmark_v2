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
BenchmarkHandleProduct-12          30000           1295360 ns/op            3354 B/op         49 allocs/op
PASS
ok      github.com/w-bt/benchmark_v2/opt_1      40.408s
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
Showing nodes accounting for 34.55s, 95.89% of 36.03s total
Dropped 104 nodes (cum <= 0.18s)
Showing top 10 nodes out of 29
      flat  flat%   sum%        cum   cum%
    15.21s 42.21% 42.21%     15.21s 42.21%  memeqbody
     7.42s 20.59% 62.81%     31.26s 86.76%  github.com/w-bt/benchmark_v2/opt_1.findProduct
     6.89s 19.12% 81.93%      8.13s 22.56%  runtime.mapiternext
     2.83s  7.85% 89.79%      2.83s  7.85%  aeshashbody
     0.50s  1.39% 91.17%      0.50s  1.39%  runtime.memequal
     0.48s  1.33% 92.51%      0.48s  1.33%  runtime.madvise
     0.45s  1.25% 93.76%      0.45s  1.25%  runtime.nanotime1
     0.35s  0.97% 94.73%      0.35s  0.97%  runtime.add
     0.21s  0.58% 95.31%      0.21s  0.58%  runtime.(*bmap).overflow (inline)
     0.21s  0.58% 95.89%      0.21s  0.58%  runtime.(*maptype).indirectkey
(pprof) top5
Showing nodes accounting for 32850ms, 91.17% of 36030ms total
Dropped 104 nodes (cum <= 180.15ms)
Showing top 5 nodes out of 29
      flat  flat%   sum%        cum   cum%
   15210ms 42.21% 42.21%    15210ms 42.21%  memeqbody
    7420ms 20.59% 62.81%    31260ms 86.76%  github.com/w-bt/benchmark_v2/opt_1.findProduct
    6890ms 19.12% 81.93%     8130ms 22.56%  runtime.mapiternext
    2830ms  7.85% 89.79%     2830ms  7.85%  aeshashbody
     500ms  1.39% 91.17%      500ms  1.39%  runtime.memequal
(pprof) top --cum
Showing nodes accounting for 32.36s, 89.81% of 36.03s total
Dropped 104 nodes (cum <= 0.18s)
Showing top 10 nodes out of 29
      flat  flat%   sum%        cum   cum%
     0.01s 0.028% 0.028%     34.63s 96.11%  github.com/w-bt/benchmark_v2/opt_1.BenchmarkHandleProduct
         0     0% 0.028%     34.63s 96.11%  testing.(*B).launch
         0     0% 0.028%     34.63s 96.11%  testing.(*B).runN
         0     0% 0.028%     34.62s 96.09%  github.com/w-bt/benchmark_v2/opt_1.handleProduct
     7.42s 20.59% 20.62%     31.26s 86.76%  github.com/w-bt/benchmark_v2/opt_1.findProduct
    15.21s 42.21% 62.84%     15.21s 42.21%  memeqbody
     6.89s 19.12% 81.96%      8.13s 22.56%  runtime.mapiternext
         0     0% 81.96%      2.84s  7.88%  net/http.Header.Set (inline)
         0     0% 81.96%      2.84s  7.88%  net/textproto.MIMEHeader.Set (inline)
     2.83s  7.85% 89.81%      2.83s  7.85%  aeshashbody
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
Total: 36.03s
ROUTINE ======================== github.com/w-bt/benchmark_v2/opt_1.handleProduct in /Users/pt.gojekindonesia/Documents/code/go/src/github.com/w-bt/benchmark_v2/opt_1/main.go
         0     34.62s (flat, cum) 96.09% of Total
         .          .     18:   http.HandleFunc("/product", handleProduct)
         .          .     19:   log.Fatal(http.ListenAndServe("127.0.0.1:1234", nil))
         .          .     20:}
         .          .     21:
         .          .     22:func handleProduct(w http.ResponseWriter, r *http.Request) {
         .       20ms     23:   code := r.FormValue("code")
         .      430ms     24:   if match, _ := regexp.MatchString(`^[A-Z]{2}[0-9]{2}$`, code); !match {
         .          .     25:           http.Error(w, "code is invalid", http.StatusBadRequest)
         .          .     26:           return
         .          .     27:   }
         .          .     28:
         .     31.26s     29:   result := findProduct(products, code)
         .          .     30:
         .          .     31:   if result.Code == "" {
         .          .     32:           http.Error(w, "data not found", http.StatusBadRequest)
         .          .     33:           return
         .          .     34:   }
         .          .     35:
         .      2.84s     36:   w.Header().Set("Content-Type", "text/html; charset=utf-8")
         .       70ms     37:   w.Write([]byte(`<font size="10">Product Code : ` + result.Code + ` Name :` + result.Name + `</font>`))
         .          .     38:}
         .          .     39:
         .          .     40:func findProduct(Products map[string]*Product, code string) *Product {
         .          .     41:   for _, item := range Products {
         .          .     42:           if code == (*item).Code {
(pprof) list findProduct
Total: 36.03s
ROUTINE ======================== github.com/w-bt/benchmark_v2/opt_1.findProduct in /Users/pt.gojekindonesia/Documents/code/go/src/github.com/w-bt/benchmark_v2/opt_1/main.go
     7.42s     31.26s (flat, cum) 86.76% of Total
         .          .     36:   w.Header().Set("Content-Type", "text/html; charset=utf-8")
         .          .     37:   w.Write([]byte(`<font size="10">Product Code : ` + result.Code + ` Name :` + result.Name + `</font>`))
         .          .     38:}
         .          .     39:
         .          .     40:func findProduct(Products map[string]*Product, code string) *Product {
     760ms      8.89s     41:   for _, item := range Products {
     6.66s     22.37s     42:           if code == (*item).Code {
         .          .     43:                   return item
         .          .     44:           }
         .          .     45:   }
         .          .     46:
         .          .     47:   return &Product{}
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
Time: Mar 14, 2021 at 4:05pm (WIB)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top --cum
Showing nodes accounting for 32MB, 29.99% of 106.70MB total
Dropped 3 nodes (cum <= 0.53MB)
Showing top 10 nodes out of 39
      flat  flat%   sum%        cum   cum%
         0     0%     0%    96.52MB 90.46%  github.com/w-bt/benchmark_v2/opt_1.BenchmarkHandleProduct
    3.50MB  3.28%  3.28%    96.52MB 90.46%  github.com/w-bt/benchmark_v2/opt_1.handleProduct
         0     0%  3.28%    96.52MB 90.46%  testing.(*B).launch
         0     0%  3.28%    96.52MB 90.46%  testing.(*B).runN
         0     0%  3.28%    92.02MB 86.24%  regexp.Compile (inline)
         0     0%  3.28%    92.02MB 86.24%  regexp.MatchString
    5.50MB  5.16%  8.44%    92.02MB 86.24%  regexp.compile
         0     0%  8.44%    31.01MB 29.06%  regexp.compileOnePass
       3MB  2.81% 11.25%       25MB 23.43%  regexp/syntax.Parse
      20MB 18.75% 29.99%       20MB 18.75%  regexp/syntax.(*parser).newRegexp (inline)
(pprof) list handleProduct
Total: 106.70MB
ROUTINE ======================== github.com/w-bt/benchmark_v2/opt_1.handleProduct in /Users/pt.gojekindonesia/Documents/code/go/src/github.com/w-bt/benchmark_v2/opt_1/main.go
    3.50MB    96.52MB (flat, cum) 90.46% of Total
         .          .     19:   log.Fatal(http.ListenAndServe("127.0.0.1:1234", nil))
         .          .     20:}
         .          .     21:
         .          .     22:func handleProduct(w http.ResponseWriter, r *http.Request) {
         .          .     23:   code := r.FormValue("code")
         .    92.02MB     24:   if match, _ := regexp.MatchString(`^[A-Z]{2}[0-9]{2}$`, code); !match {
         .          .     25:           http.Error(w, "code is invalid", http.StatusBadRequest)
         .          .     26:           return
         .          .     27:   }
         .          .     28:
         .          .     29:   result := findProduct(products, code)
         .          .     30:
         .          .     31:   if result.Code == "" {
         .          .     32:           http.Error(w, "data not found", http.StatusBadRequest)
         .          .     33:           return
         .          .     34:   }
         .          .     35:
         .        1MB     36:   w.Header().Set("Content-Type", "text/html; charset=utf-8")
    3.50MB     3.50MB     37:   w.Write([]byte(`<font size="10">Product Code : ` + result.Code + ` Name :` + result.Name + `</font>`))
         .          .     38:}
         .          .     39:
         .          .     40:func findProduct(Products map[string]*Product, code string) *Product {
         .          .     41:   for _, item := range Products {
         .          .     42:           if code == (*item).Code {
(pprof) list MatchString
Total: 106.70MB
ROUTINE ======================== regexp.MatchString in /usr/local/Cellar/go/1.14.2/libexec/src/regexp/regexp.go
         0    92.02MB (flat, cum) 86.24% of Total
         .          .    526:
         .          .    527:// MatchString reports whether the string s
         .          .    528:// contains any match of the regular expression pattern.
         .          .    529:// More complicated queries need to use Compile and the full Regexp interface.
         .          .    530:func MatchString(pattern string, s string) (matched bool, err error) {
         .    92.02MB    531:   re, err := Compile(pattern)
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
BenchmarkHandleProduct-12          30000              4529 ns/op            3353 B/op         49 allocs/op
PASS
ok      github.com/w-bt/benchmark/opt_2 1.618s
```
```sh
$ benchcmp ../opt_1/prof.txt prof.txt     
benchmark                     old ns/op     new ns/op     delta
BenchmarkHandleProduct-12     1295360       4529          -99.65%

benchmark                     old allocs     new allocs     delta
BenchmarkHandleProduct-12     49             49             +0.00%

benchmark                     old bytes     new bytes     delta
BenchmarkHandleProduct-12     3354          3353          -0.03%
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
BenchmarkHandleProduct-12          30000               386 ns/op             144 B/op          3 allocs/op
PASS
ok      github.com/w-bt/benchmark/opt_3 1.485s
```
```sh
$ benchcmp ../opt_2/prof.txt prof.txt
benchmark                     old ns/op     new ns/op     delta
BenchmarkHandleProduct-12     4529          386           -91.48%

benchmark                     old allocs     new allocs     delta
BenchmarkHandleProduct-12     49             3              -93.88%

benchmark                     old bytes     new bytes     delta
BenchmarkHandleProduct-12     3353          144           -95.71%
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
BenchmarkHandleProduct-12          30000               300 ns/op              16 B/op          1 allocs/op
PASS
ok      github.com/w-bt/benchmark/opt_4 1.221s
```
```sh
$ benchcmp ../opt_3/prof.txt prof.txt
benchmark                     old ns/op     new ns/op     delta
BenchmarkHandleProduct-12     386           300           -22.28%

benchmark                     old allocs     new allocs     delta
BenchmarkHandleProduct-12     3              1              -66.67%

benchmark                     old bytes     new bytes     delta
BenchmarkHandleProduct-12     144           16            -88.89%
```
