package main

import (
	"fmt"

	"github.com/jinzhu/inflection"
)

func main() {
	p := "id"
	ps := inflection.Plural(p)
	fmt.Println(ps) // echo "ids"
	b := "bus"
	bs := inflection.Plural(b)
	fmt.Println(bs) // echo "buses"
}
