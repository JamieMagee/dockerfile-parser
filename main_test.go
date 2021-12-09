package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestParseBaseImages(t *testing.T) {
	var tests = []struct {
		file string
		want []BaseImage
	}{
		{
			file: "Dockerfile",
			want: []BaseImage{
				{
					Registry: "docker.io",
					Image:    "library/ubuntu",
					Stage:    0,
				},
			},
		},
	}

	for _, tt := range tests {
		reader := loadTestFile(tt.file)
		t.Run(fmt.Sprintf("%v", tt.file), func(t *testing.T) {
			result := parseBaseImages(reader)
			if !reflect.DeepEqual(tt.want, result) {
				t.Errorf("expected %v, got %v", tt.want, result)
			}
		})
	}

}

func loadTestFile(file string) *bytes.Reader {
	dat, _ := os.ReadFile(path.Join("fixtures", file))
	reader := bytes.NewReader(dat)
	return reader
}
