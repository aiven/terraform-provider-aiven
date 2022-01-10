package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

func init() {
	flag.Parse()
}

var (
	terraformVersion = version.Must(version.NewVersion("1.0.6"))
)

func main() {
	ctx := context.Background()

	log.Println("[INFO] installing terraform")

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: terraformVersion,
	}

	execPath, err := installer.Install(ctx)
	if err != nil {
		log.Fatalf("[ERROR] unable to intsall terraform: %s", err)
	}

	tmpdir := os.TempDir()
	defer func() { _ = os.RemoveAll(tmpdir) }()

	tf, err := tfexec.NewTerraform(tmpdir, execPath)
	if err != nil {
		log.Fatalf("[ERROR] unable to create terraform handle: %s", err)
	}

	log.Println("[INFO] installed terraform, formatting files")

	for _, path := range flag.Args() {
		log.Println("[INFO] formatting ", path)
		if err = reformatFileAtPath(ctx, path, tf); err != nil {
			log.Println("[ERROR]: unable to format ", path, ": ", err)
		}
	}

	log.Println("[INFO] formatted all files")
}

func reformatFileAtPath(ctx context.Context, path string, tf *tfexec.Terraform) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("unable to open file at path '%s': %w", path, err)
	}
	defer func() { _ = file.Close() }()

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, "src.go", file, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("unable to parse file at path '%s': %w", path, err)
	}

	var (
		hadFmtErr bool
		stack     []ast.Node
	)
	ast.Inspect(parsed, func(n ast.Node) bool {
		defer func() {
			if n == nil {
				// Done with node's children. Pop.
				stack = stack[:len(stack)-1]
			} else {
				// Push the current node for children.
				stack = append(stack, n)
			}
		}()

		if n == nil || len(stack) == 0 {
			return true
		}

		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING || !isTerraformManifest(lit.Value) {
			return true
		}

		indentation := getIndentation(stack, fset)
		unquoted := strings.Trim(lit.Value, "`")
		stripped := strings.ReplaceAll(unquoted, "\t", "")
		if formatted, err := tf.FormatString(ctx, stripped); err != nil {
			log.Println("[ERROR]: unable to format node ", stripped, ": ", err)
			hadFmtErr = true
		} else {
			lit.Value = "`" + strings.TrimRightFunc(strings.ReplaceAll(formatted, "\n", "\n"+strings.Repeat("\t", indentation)), unicode.IsSpace) + "`"
		}
		return true

	})

	if hadFmtErr {
		return fmt.Errorf("unable to format file at path '%s'", path)
	}

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("unable to truncate file at path '%s': %w", path, err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("unable to seek beginning of file at path '%s': %w", path, err)
	}

	return format.Node(file, fset, parsed)
}

func isTerraformManifest(s string) bool {
	return strings.Contains(s, "aiven_") &&
		strings.Contains(s, "{") &&
		strings.Contains(s, "}")
}

func getIndentation(stack []ast.Node, fset *token.FileSet) int {
	if ret := lastReturnStatement(stack); ret != nil {
		return fset.Position(ret.Pos()).Column
	}
	if ass := lastAssignmentStatement(stack); ass != nil {
		return fset.Position(ass.Pos()).Column
	}
	return 0
}

func lastReturnStatement(stack []ast.Node) ast.Node {
	for i := len(stack) - 1; i != 0; i-- {
		switch stack[i].(type) {
		case *ast.ReturnStmt:
			return stack[i]
		default:
			continue
		}
	}
	return nil
}

func lastAssignmentStatement(stack []ast.Node) ast.Node {
	for i := len(stack) - 1; i != 0; i-- {
		switch stack[i].(type) {
		case *ast.AssignStmt:
			return stack[i]
		default:
			continue
		}
	}
	return nil
}
