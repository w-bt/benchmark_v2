package main

import (
	"fmt"
	"github.com/go-redis/redis/v8"
)

var (
	rdb *redis.Client
)

func GenerateProduct() {
	products = make(map[string]*Product)
	//result, err := rdb.Get(context.Background(), "product:test").Result()
	//if err != nil && err != redis.Nil {
	//	panic(err)
	//}
	//
	//if result != "" {
	//	log.Println("load products from cache")
	//	err = json.Unmarshal([]byte(result), &products)
	//	if err != nil {
	//		panic(err)
	//	}
	//	return
	//}

	generateAndSetProductToRedis()

	return
}

func generateAndSetProductToRedis() {
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

	//jsonByte, err := json.Marshal(products)
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = rdb.Set(context.Background(), "product:test", string(jsonByte), time.Minute*60).Err()
	//if err != nil {
	//	panic(err)
	//}
}
