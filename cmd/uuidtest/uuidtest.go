package main

import (
	"fmt"
	uuid "github.com/google/uuid"
)

func testv1() {
	id := uuid.New()
	fmt.Printf("%s %s\n", id, id.Version().String())
}

func main() {
	fmt.Println("Hello, 世界")
	for i := 0; i < 5; i++ {
		testv1()
	}
}
