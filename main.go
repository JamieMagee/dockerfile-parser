package main

import (
	"dockerfile-parser/dockerfile"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	df := loadFile(os.Args[1])
	result, err := df.FindImages()
	if err != nil {
		panic(err)
	}

	resultBytes, _ := json.Marshal(result)
	fmt.Println(string(resultBytes))
}

func loadFile(p string) dockerfile.Dockerfile {
	dat, err := os.ReadFile(p)
	if err != nil {
		panic(err)
	}

	return dockerfile.Dockerfile(dat)
}
