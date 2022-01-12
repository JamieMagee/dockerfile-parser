package main

import (
	"dockerfile-parser/dockerfile"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestParseBaseImages(t *testing.T) {
	var tests = []struct {
		file string
		want []dockerfile.BaseImage
	}{
		{
			file: "Dockerfile",
			want: []dockerfile.BaseImage{
				{
					Registry: "docker.io",
					Image:    "library/ubuntu",
					Tag:      "latest",
				},
			},
		},
		{
			file: "Dockerfile.2",
			want: []dockerfile.BaseImage{
				{
					Registry: "mcr.microsoft.com",
					Image:    "dotnet/sdk",
					Tag:      "6.0",
				},
				{
					Image: "scratch",
				},
				{
					Registry: "mcr.microsoft.com",
					Image:    "dotnet/sdk",
					Tag:      "6.0",
				},
			},
		},
	}

	for _, tt := range tests {
		df := loadTestFile(tt.file)
		t.Run(fmt.Sprintf("%v", tt.file), func(t *testing.T) {
			result, _ := df.FindImages()
			if !reflect.DeepEqual(tt.want, result) {
				t.Errorf("expected %v, got %v", tt.want, result)
			}
		})
	}

}

func loadTestFile(file string) dockerfile.Dockerfile {
	dat, _ := os.ReadFile(path.Join("fixtures", file))
	return dockerfile.Dockerfile(dat)
}
