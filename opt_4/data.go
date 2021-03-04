package main

import (
	"fmt"
)

func GenerateProduct() {
	products = make(map[string]*Product)
	ch := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	no := 0
	for _, c := range ch {
		for _, c2 := range ch {
			for i := 0; i < 10; i++ {
				for j := 0; j < 10; j++ {
					p := Product{}
					p.Code = fmt.Sprintf("%s%s%d%d", c, c2, i, j)
					p.Name = fmt.Sprintf("Product %d", no)
					products[p.Code] = &p
					no++
				}
			}
		}
	}

	return
}
