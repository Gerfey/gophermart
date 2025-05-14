// Multichecker включает в себя следующие анализаторы:
//   - Стандартные анализаторы из пакета golang.org/x/tools/go/analysis/passes
//   - Все анализаторы класса SA из пакета staticcheck.io
//   - Анализаторы из других классов пакета staticcheck.io (ST1000, unused)
//   - Публичные анализаторы (errcheck)
//   - Собственный анализатор noexitinmain, запрещающий использование os.Exit в функции main пакета main
//
// Запуск:
//
//	go run cmd/staticlint/main.go ./...
//
// Для запуска конкретного анализатора:
//
//	go run cmd/staticlint/main.go -noexitinmain ./...
//
// Для получения списка всех доступных анализаторов:
//
//	go run cmd/staticlint/main.go -help
package main

import (
	"go/ast"
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

// NoExitAnalyzer - анализатор, который запрещает использование os.Exit в функции main пакета main.
// Этот анализатор проверяет, что в функции main пакета main не используется прямой вызов os.Exit.
var NoExitAnalyzer = &analysis.Analyzer{
	Name: "noexitinmain",
	Doc:  "Запрещает использование прямого вызова os.Exit в функции main пакета main",
	Run:  runNoExitAnalyzer,
}

// runNoExitAnalyzer реализует логику анализатора NoExitAnalyzer.
// Функция проверяет все файлы в пакете main и ищет вызовы os.Exit в функции main.
// Если такие вызовы найдены, анализатор сообщает об ошибке.
func runNoExitAnalyzer(pass *analysis.Pass) (interface{}, error) {
	if len(pass.Files) == 0 {
		return nil, nil
	}

	fileName := pass.Fset.File(pass.Files[0].Pos()).Name()

	if strings.Contains(fileName, "go-build") || strings.Contains(fileName, "/Caches/") {
		return nil, nil
	}

	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" {
				continue
			}

			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok {
							if ident.Name == "os" && sel.Sel.Name == "Exit" {
								pass.Reportf(call.Pos(), "использование os.Exit в функции main запрещено")
							}
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}

func main() {
	standardAnalyzers := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}

	saAnalyzers := make([]*analysis.Analyzer, 0)
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" {
			saAnalyzers = append(saAnalyzers, v.Analyzer)
		}
	}

	var otherAnalyzers []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name == "ST1005" {
			otherAnalyzers = append(otherAnalyzers, v.Analyzer)
			break
		}
	}

	publicAnalyzers := []*analysis.Analyzer{
		errcheck.Analyzer,
	}

	allAnalyzers := append(standardAnalyzers, saAnalyzers...)
	allAnalyzers = append(allAnalyzers, otherAnalyzers...)
	allAnalyzers = append(allAnalyzers, publicAnalyzers...)
	allAnalyzers = append(allAnalyzers, NoExitAnalyzer)

	multichecker.Main(allAnalyzers...)
}
