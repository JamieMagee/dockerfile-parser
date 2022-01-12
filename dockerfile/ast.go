package dockerfile

import (
	"bytes"
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"
	"github.com/pkg/errors"
	"io"
	"strings"
)

type AST struct {
	directives map[string]dockerfile2llb.Directive
	result     *parser.Result
}

func ParseAST(df Dockerfile) (AST, error) {
	result, err := parser.Parse(newReader(df))
	if err != nil {
		return AST{}, errors.Wrap(err, "dockerfile.ParseAST")
	}

	return AST{
		directives: dockerfile2llb.ParseDirectives(newReader(df)),
		result:     result,
	}, nil
}

func (a AST) traverseImageRefs(visitor func(node *parser.Node, ref reference.Named) reference.Named) error {
	metaArgs := append([]instructions.ArgCommand(nil))
	stageNames := map[string]string{}
	shlex := shell.NewLex(a.result.EscapeToken)

	return a.Traverse(func(node *parser.Node) error {
		switch strings.ToLower(node.Value) {
		case command.Arg:
			inst, err := instructions.ParseInstruction(node)
			if err != nil {
				return nil // ignore parsing error
			}

			argCmd, ok := inst.(*instructions.ArgCommand)
			if !ok {
				return nil
			}

			// args within the Dockerfile are prepended because they provide defaults that are overridden by actual args
			metaArgs = append([]instructions.ArgCommand{*argCmd}, metaArgs...)

		case command.From:
			baseName, stageName := a.extractBaseNameInFromCommand(node, shlex, metaArgs)
			if baseName == "" {
				return nil // ignore parsing error
			}

			ref, err := reference.ParseNormalizedNamed(baseName)
			if err != nil {
				return nil // drop the error, we don't care about malformed images
			}
			stageNames[stageName] = baseName
			newRef := visitor(node, ref)
			if newRef != nil {
				node.Next.Value = reference.FamiliarString(newRef)
			}

		case command.Copy:
			if len(node.Flags) == 0 {
				return nil
			}

			inst, err := instructions.ParseInstruction(node)
			if err != nil {
				return nil // ignore parsing error
			}

			copyCmd, ok := inst.(*instructions.CopyCommand)
			if !ok {
				return nil
			}

			ref, err := reference.ParseNormalizedNamed(copyCmd.From)
			if err != nil {
				return nil // drop the error, we don't care about malformed images
			}

			for _, flag := range node.Flags {
				if strings.HasPrefix(flag, "--from=") {
					fromRef := strings.TrimPrefix(flag, "--from=")
					if x, found := stageNames[fromRef]; found {
						ref, err = reference.ParseNormalizedNamed(x)
						if err != nil {
							panic(nil)
						}
					}
				}
			}

			visitor(node, ref)
		}

		return nil
	})
}

func (a AST) extractBaseNameInFromCommand(node *parser.Node, shlex *shell.Lex, metaArgs []instructions.ArgCommand) (imageName string, stageName string) {
	if node.Next == nil {
		return "", ""
	}

	inst, err := instructions.ParseInstruction(node)
	if err != nil {
		return node.Next.Value, "" // if there's a parsing error, fallback to the first arg
	}

	fromInst, ok := inst.(*instructions.Stage)
	if !ok || fromInst.BaseName == "" {
		return "", ""
	}

	// The base image name may have ARG expansions in it. Do the default
	// substitution.
	argsMap := fakeArgsMap(shlex, metaArgs)
	baseName, err := shlex.ProcessWordWithMap(fromInst.BaseName, argsMap)
	if err != nil {
		// If anything fails, just use the hard-coded BaseName.
		return fromInst.BaseName, ""
	}
	return baseName, fromInst.Name

}

func (a AST) Traverse(visit func(*parser.Node) error) error {
	return a.traverseNode(a.result.AST, visit)
}

func (a AST) traverseNode(node *parser.Node, visit func(*parser.Node) error) error {
	for _, c := range node.Children {
		err := a.traverseNode(c, visit)
		if err != nil {
			return err
		}
	}
	return visit(node)
}

func newReader(df Dockerfile) io.Reader {
	return bytes.NewBufferString(string(df))
}

// Loosely adapted from the buildkit code for turning args into a map.
// Iterate through them and do substitutions in order.
func fakeArgsMap(shlex *shell.Lex, args []instructions.ArgCommand) map[string]string {
	m := make(map[string]string)
	for _, argCmd := range args {
		val := ""
		for _, a := range argCmd.Args {
			if a.Value != nil {
				val, _ = shlex.ProcessWordWithMap(*(a.Value), m)
			}
			m[a.Key] = val
		}
	}
	return m
}
