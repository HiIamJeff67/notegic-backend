package architecture_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const runtimeImportPrefix = "github.com/HiIamJeff67/notegic-backend/runtimes/"

func TestNoCrossServiceSourceImports(t *testing.T) {
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve the repository root")
	}

	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFilePath), "..", ".."))
	serviceRoot := filepath.Join(repositoryRoot, "runtimes")
	serviceEntries, err := filepath.Glob(filepath.Join(serviceRoot, "*"))
	if err != nil {
		t.Fatalf("failed to list service directories: %v", err)
	}

	for _, servicePath := range serviceEntries {
		serviceName := filepath.Base(servicePath)
		if err := filepath.WalkDir(servicePath, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}

			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, importSpec := range file.Imports {
				importPath, err := strconv.Unquote(importSpec.Path.Value)
				if err != nil {
					return err
				}
				if !strings.HasPrefix(importPath, runtimeImportPrefix) {
					continue
				}

				importedServiceName, _, _ := strings.Cut(
					strings.TrimPrefix(importPath, runtimeImportPrefix),
					"/",
				)
				if importedServiceName == "" || importedServiceName == serviceName {
					continue
				}
				if !strings.Contains(importedServiceName, "core") && !strings.Contains(serviceName, "core") {
					continue
				}
				if importedServiceName != serviceName {
					return fmt.Errorf(
						"%s in service %s imports service %s",
						path,
						serviceName,
						importedServiceName,
					)
				}
			}

			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSharedDoesNotDependOnApplicationPackages(t *testing.T) {
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve the repository root")
	}

	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFilePath), "..", ".."))
	sharedRoot := filepath.Join(repositoryRoot, "shared")
	if err := filepath.WalkDir(sharedRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, importSpec := range file.Imports {
			importPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(importPath, runtimeImportPrefix) {
				return fmt.Errorf("%s imports application package %s", path, importPath)
			}
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestExceptionsDoesNotDependOnTransportPackages(t *testing.T) {
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve the repository root")
	}

	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFilePath), "..", ".."))
	exceptionRoot := filepath.Join(repositoryRoot, "contracts", "types", "exceptions")
	if err := filepath.WalkDir(exceptionRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, importSpec := range file.Imports {
			importPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(importPath, "github.com/99designs/gqlgen/") ||
				strings.HasPrefix(importPath, "github.com/vektah/gqlparser/") {
				return fmt.Errorf("%s imports transport package %s", path, importPath)
			}
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAPIDataDoesNotUseLegacyOwnershipPaths(t *testing.T) {
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve the repository root")
	}

	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFilePath), "..", ".."))
	legacyRoots := []string{
		filepath.Join(repositoryRoot, "runtimes", "models"),
		filepath.Join(repositoryRoot, "runtimes", "options"),
		filepath.Join(repositoryRoot, "runtimes", "modules"),
	}
	for _, legacyRoot := range legacyRoots {
		if _, err := os.Stat(legacyRoot); err == nil {
			t.Fatalf("legacy ownership path still exists: %s", legacyRoot)
		} else if !os.IsNotExist(err) {
			t.Fatalf("inspect %s: %v", legacyRoot, err)
		}
	}
}

func TestLegacyGatewayBinderPackageDoesNotExist(t *testing.T) {
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve the repository root")
	}

	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFilePath), "..", ".."))
	binderRoot := filepath.Join(repositoryRoot, "runtimes", "binders")
	if _, err := os.Stat(binderRoot); err == nil {
		t.Fatalf("legacy Gateway binder package still exists: %s", binderRoot)
	} else if !os.IsNotExist(err) {
		t.Fatalf("inspect %s: %v", binderRoot, err)
	}
}
