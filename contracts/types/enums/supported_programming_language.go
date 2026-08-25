package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type SupportedProgrammingLanguage string

const (
	SupportedProgrammingLanguage_Abap         SupportedProgrammingLanguage = "abap"
	SupportedProgrammingLanguage_Arduino      SupportedProgrammingLanguage = "arduino"
	SupportedProgrammingLanguage_Bash         SupportedProgrammingLanguage = "bash"
	SupportedProgrammingLanguage_Basic        SupportedProgrammingLanguage = "basic"
	SupportedProgrammingLanguage_C            SupportedProgrammingLanguage = "c"
	SupportedProgrammingLanguage_Clojure      SupportedProgrammingLanguage = "clojure"
	SupportedProgrammingLanguage_CoffeeScript SupportedProgrammingLanguage = "coffeescript"
	SupportedProgrammingLanguage_CPP          SupportedProgrammingLanguage = "cpp"
	SupportedProgrammingLanguage_CSharp       SupportedProgrammingLanguage = "csharp"
	SupportedProgrammingLanguage_CSS          SupportedProgrammingLanguage = "css"
	SupportedProgrammingLanguage_Dart         SupportedProgrammingLanguage = "dart"
	SupportedProgrammingLanguage_Diff         SupportedProgrammingLanguage = "diff"
	SupportedProgrammingLanguage_Docker       SupportedProgrammingLanguage = "docker"
	SupportedProgrammingLanguage_Elixir       SupportedProgrammingLanguage = "elixir"
	SupportedProgrammingLanguage_Elm          SupportedProgrammingLanguage = "elm"
	SupportedProgrammingLanguage_Erlang       SupportedProgrammingLanguage = "erlang"
	SupportedProgrammingLanguage_Flow         SupportedProgrammingLanguage = "flow"
	SupportedProgrammingLanguage_Fortran      SupportedProgrammingLanguage = "fortran"
	SupportedProgrammingLanguage_FSharp       SupportedProgrammingLanguage = "fsharp"
	SupportedProgrammingLanguage_Gherkin      SupportedProgrammingLanguage = "gherkin"
	SupportedProgrammingLanguage_Glsl         SupportedProgrammingLanguage = "glsl"
	SupportedProgrammingLanguage_Go           SupportedProgrammingLanguage = "go"
	SupportedProgrammingLanguage_GraphQL      SupportedProgrammingLanguage = "graphql"
	SupportedProgrammingLanguage_Groovy       SupportedProgrammingLanguage = "groovy"
	SupportedProgrammingLanguage_Haskell      SupportedProgrammingLanguage = "haskell"
	SupportedProgrammingLanguage_HTML         SupportedProgrammingLanguage = "html"
	SupportedProgrammingLanguage_Java         SupportedProgrammingLanguage = "java"
	SupportedProgrammingLanguage_JavaScript   SupportedProgrammingLanguage = "javascript"
	SupportedProgrammingLanguage_JSON         SupportedProgrammingLanguage = "json"
	SupportedProgrammingLanguage_Julia        SupportedProgrammingLanguage = "julia"
	SupportedProgrammingLanguage_Kotlin       SupportedProgrammingLanguage = "kotlin"
	SupportedProgrammingLanguage_Latex        SupportedProgrammingLanguage = "latex"
	SupportedProgrammingLanguage_Less         SupportedProgrammingLanguage = "less"
	SupportedProgrammingLanguage_Lisp         SupportedProgrammingLanguage = "lisp"
	SupportedProgrammingLanguage_LiveScript   SupportedProgrammingLanguage = "livescript"
	SupportedProgrammingLanguage_Lua          SupportedProgrammingLanguage = "lua"
	SupportedProgrammingLanguage_Makefile     SupportedProgrammingLanguage = "makefile"
	SupportedProgrammingLanguage_Markdown     SupportedProgrammingLanguage = "markdown"
	SupportedProgrammingLanguage_Markup       SupportedProgrammingLanguage = "markup"
	SupportedProgrammingLanguage_Matlab       SupportedProgrammingLanguage = "matlab"
	SupportedProgrammingLanguage_Nix          SupportedProgrammingLanguage = "nix"
	SupportedProgrammingLanguage_ObjectC      SupportedProgrammingLanguage = "objective-c"
	SupportedProgrammingLanguage_OCAML        SupportedProgrammingLanguage = "ocaml"
	SupportedProgrammingLanguage_Pascal       SupportedProgrammingLanguage = "pascal"
	SupportedProgrammingLanguage_Perl         SupportedProgrammingLanguage = "perl"
	SupportedProgrammingLanguage_PHP          SupportedProgrammingLanguage = "php"
	SupportedProgrammingLanguage_Plaintext    SupportedProgrammingLanguage = "plaintext"
	SupportedProgrammingLanguage_Text         SupportedProgrammingLanguage = "text"
	SupportedProgrammingLanguage_Txt          SupportedProgrammingLanguage = "txt"
	SupportedProgrammingLanguage_Powershell   SupportedProgrammingLanguage = "powershell"
	SupportedProgrammingLanguage_Prolog       SupportedProgrammingLanguage = "prolog"
	SupportedProgrammingLanguage_Protobuf     SupportedProgrammingLanguage = "protobuf"
	SupportedProgrammingLanguage_Python       SupportedProgrammingLanguage = "python"
	SupportedProgrammingLanguage_R            SupportedProgrammingLanguage = "r"
	SupportedProgrammingLanguage_Reason       SupportedProgrammingLanguage = "reason"
	SupportedProgrammingLanguage_Ruby         SupportedProgrammingLanguage = "ruby"
	SupportedProgrammingLanguage_Rust         SupportedProgrammingLanguage = "rust"
	SupportedProgrammingLanguage_SCSS         SupportedProgrammingLanguage = "scss"
	SupportedProgrammingLanguage_Shell        SupportedProgrammingLanguage = "shell"
	SupportedProgrammingLanguage_SQL          SupportedProgrammingLanguage = "sql"
	SupportedProgrammingLanguage_Swift        SupportedProgrammingLanguage = "swift"
	SupportedProgrammingLanguage_TypeScript   SupportedProgrammingLanguage = "typescript"
	SupportedProgrammingLanguage_VBDotNet     SupportedProgrammingLanguage = "vb.net"
	SupportedProgrammingLanguage_Verilog      SupportedProgrammingLanguage = "verilog"
	SupportedProgrammingLanguage_VHDL         SupportedProgrammingLanguage = "vhdl"
	SupportedProgrammingLanguage_VisualBasic  SupportedProgrammingLanguage = "visual-basic"
	SupportedProgrammingLanguage_WebAssembly  SupportedProgrammingLanguage = "webassembly"
	SupportedProgrammingLanguage_XML          SupportedProgrammingLanguage = "xml"
	SupportedProgrammingLanguage_YAML         SupportedProgrammingLanguage = "yaml"
)

var AllSupportedProgrammingLanguages = []SupportedProgrammingLanguage{
	SupportedProgrammingLanguage_Abap,
	SupportedProgrammingLanguage_Arduino,
	SupportedProgrammingLanguage_Bash,
	SupportedProgrammingLanguage_Basic,
	SupportedProgrammingLanguage_C,
	SupportedProgrammingLanguage_Clojure,
	SupportedProgrammingLanguage_CoffeeScript,
	SupportedProgrammingLanguage_CPP,
	SupportedProgrammingLanguage_CSharp,
	SupportedProgrammingLanguage_CSS,
	SupportedProgrammingLanguage_Dart,
	SupportedProgrammingLanguage_Diff,
	SupportedProgrammingLanguage_Docker,
	SupportedProgrammingLanguage_Elixir,
	SupportedProgrammingLanguage_Elm,
	SupportedProgrammingLanguage_Erlang,
	SupportedProgrammingLanguage_Flow,
	SupportedProgrammingLanguage_Fortran,
	SupportedProgrammingLanguage_FSharp,
	SupportedProgrammingLanguage_Gherkin,
	SupportedProgrammingLanguage_Glsl,
	SupportedProgrammingLanguage_Go,
	SupportedProgrammingLanguage_GraphQL,
	SupportedProgrammingLanguage_Groovy,
	SupportedProgrammingLanguage_Haskell,
	SupportedProgrammingLanguage_HTML,
	SupportedProgrammingLanguage_Java,
	SupportedProgrammingLanguage_JavaScript,
	SupportedProgrammingLanguage_JSON,
	SupportedProgrammingLanguage_Julia,
	SupportedProgrammingLanguage_Kotlin,
	SupportedProgrammingLanguage_Latex,
	SupportedProgrammingLanguage_Less,
	SupportedProgrammingLanguage_Lisp,
	SupportedProgrammingLanguage_LiveScript,
	SupportedProgrammingLanguage_Lua,
	SupportedProgrammingLanguage_Makefile,
	SupportedProgrammingLanguage_Markdown,
	SupportedProgrammingLanguage_Markup,
	SupportedProgrammingLanguage_Matlab,
	SupportedProgrammingLanguage_Nix,
	SupportedProgrammingLanguage_ObjectC,
	SupportedProgrammingLanguage_OCAML,
	SupportedProgrammingLanguage_Pascal,
	SupportedProgrammingLanguage_Perl,
	SupportedProgrammingLanguage_PHP,
	SupportedProgrammingLanguage_Plaintext,
	SupportedProgrammingLanguage_Text,
	SupportedProgrammingLanguage_Txt,
	SupportedProgrammingLanguage_Powershell,
	SupportedProgrammingLanguage_Prolog,
	SupportedProgrammingLanguage_Protobuf,
	SupportedProgrammingLanguage_Python,
	SupportedProgrammingLanguage_R,
	SupportedProgrammingLanguage_Reason,
	SupportedProgrammingLanguage_Ruby,
	SupportedProgrammingLanguage_Rust,
	SupportedProgrammingLanguage_SCSS,
	SupportedProgrammingLanguage_Shell,
	SupportedProgrammingLanguage_SQL,
	SupportedProgrammingLanguage_Swift,
	SupportedProgrammingLanguage_TypeScript,
	SupportedProgrammingLanguage_VBDotNet,
	SupportedProgrammingLanguage_Verilog,
	SupportedProgrammingLanguage_VHDL,
	SupportedProgrammingLanguage_VisualBasic,
	SupportedProgrammingLanguage_WebAssembly,
	SupportedProgrammingLanguage_XML,
	SupportedProgrammingLanguage_YAML,
}

var AllSupportedProgrammingLanguageStrings = []string{
	"abap",
	"arduino",
	"bash",
	"basic",
	"c",
	"clojure",
	"coffeescript",
	"cpp",
	"csharp",
	"css",
	"dart",
	"diff",
	"docker",
	"elixir",
	"elm",
	"erlang",
	"flow",
	"fortran",
	"fsharp",
	"gherkin",
	"glsl",
	"go",
	"graphql",
	"groovy",
	"haskell",
	"html",
	"java",
	"javascript",
	"json",
	"julia",
	"kotlin",
	"latex",
	"less",
	"lisp",
	"livescript",
	"lua",
	"makefile",
	"markdown",
	"markup",
	"matlab",
	"nix",
	"objective-c",
	"ocaml",
	"pascal",
	"perl",
	"php",
	"plaintext",
	"text",
	"txt",
	"powershell",
	"prolog",
	"protobuf",
	"python",
	"r",
	"reason",
	"ruby",
	"rust",
	"sass",
	"scala",
	"scheme",
	"scss",
	"shell",
	"sql",
	"swift",
	"typescript",
	"vb.net",
	"verilog",
	"vhdl",
	"visual-basic",
	"webassembly",
	"xml",
	"yaml",
}

func (value SupportedProgrammingLanguage) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *SupportedProgrammingLanguage) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = SupportedProgrammingLanguage(string(v))
		return nil
	case string:
		*value = SupportedProgrammingLanguage(v)
		return nil
	}
	return scanError(raw, value)
}

func (value SupportedProgrammingLanguage) Value() (driver.Value, error) {
	return string(value), nil
}

func (value SupportedProgrammingLanguage) String() string {
	return string(value)
}

func (value *SupportedProgrammingLanguage) IsValidEnum() bool {
	return slices.Contains(AllSupportedProgrammingLanguages, *value)
}

func ConvertStringToSupportedProgrammingLanguage(enumString string) (*SupportedProgrammingLanguage, error) {
	for _, enumValue := range AllSupportedProgrammingLanguages {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid SupportedProgrammingLanguage: %s", enumString)
}
