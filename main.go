package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/distribution/distribution/reference"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

type BaseImage struct {
	Registry string

	Image string

	Tag string

	Stage int
}

func main() {
	reader := loadFile(os.Args[1])
	baseImages := parseBaseImages(reader)
	baseImagesBytes, _ := json.Marshal(baseImages)
	fmt.Println(string(baseImagesBytes))
}

func loadFile(p string) *bytes.Reader {
	dat, err := os.ReadFile(p)
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(dat)
}

func parseBaseImages(reader *bytes.Reader) []BaseImage {
	res, err := parser.Parse(reader)
	if err != nil {
		panic(err)
	}

	var baseImages []BaseImage
	var stages []*instructions.Stage

	for _, child := range res.AST.Children {
		instr, err := instructions.ParseInstruction(child)
		if err != nil {
			panic(err)
		}

		if child.Value != "FROM" {
			continue
		}

		stage, ok := instr.(*instructions.Stage)
		if ok {
			stages = append(stages, stage)
		}

		for n := child.Next; n != nil; n = n.Next {
			canonical, err := reference.ParseAnyReference(n.Value)
			if err != nil {
				panic(err)
			}

			named, err := reference.ParseNamed(canonical.String())
			if err != nil {
				panic(err)
			}

			registry := reference.Domain(named)
			image := reference.Path(named)

			baseImage := BaseImage{
				Registry: registry,
				Image:    image,
				Stage:    currentStage(stages),
			}
			baseImages = append(baseImages, baseImage)
		}
	}
	return baseImages

}

func currentStage(stages []*instructions.Stage) int {
	if len(stages) == 0 {
		return 0
	}

	return len(stages) - 1
}
