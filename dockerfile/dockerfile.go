package dockerfile

import (
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

type Dockerfile string

type BaseImage struct {
	Registry string `json:"registry,omitempty"`
	Image    string `json:"image,omitempty"`
	Tag      string `json:"tag,omitempty"`
}

func (d Dockerfile) FindImages() ([]BaseImage, error) {
	var result []reference.Named
	ast, err := ParseAST(d)
	if err != nil {
		return nil, err
	}

	err = ast.traverseImageRefs(func(node *parser.Node, ref reference.Named) reference.Named {
		result = append(result, ref)
		return nil
	})
	if err != nil {
		return nil, err
	}

	var baseImages []BaseImage
	for _, image := range result {
		if tagged, ok := image.(reference.Tagged); ok {
			baseImages = append(baseImages, BaseImage{
				Registry: reference.Domain(image),
				Image:    reference.Path(image),
				Tag:      tagged.Tag(),
			})
		} else if reference.Path(image) == "library/scratch" {
			baseImages = append(baseImages, BaseImage{
				Image: "scratch",
			})
		} else {
			baseImages = append(baseImages, BaseImage{
				Registry: reference.Domain(image),
				Image:    reference.Path(image),
			})
		}
	}

	return baseImages, nil
}
