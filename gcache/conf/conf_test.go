package conf

import (
	"fmt"
	"testing"
)

func TestInit(t *testing.T) {
	Init("./conf.json")
	fmt.Printf("%v", GConfig)
}
