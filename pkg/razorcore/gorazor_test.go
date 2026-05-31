package razorcore

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testdata = "testdata"

func TestCap(t *testing.T) {
	if Capitalize("") != "" {
		t.Error()
	}
	if Capitalize("hello") != "Hello" {
		t.Error()
	}
	if Capitalize("0hello") != "0hello" {
		t.Error()
	}
}

func TestLayManager(t *testing.T) {
	SetLayout("hello", []string{"this", "is", "good"})
	SetLayout("world", []string{"funny"})
	if len(LayoutArgs("hello")) != 3 {
		t.Error()
	}
	if len(LayoutArgs("world")) != 1 {
		t.Error()
	}
	if len(LayoutArgs("NO")) != 0 {
		t.Error()
	}
}

func TestLexer1(t *testing.T) {
	text := "case do func var switch"
	lex := &Lexer{text, Tests}
	res, err := lex.Scan()
	if err != nil {
		t.Error(err)
	}
	if len(res) != 10 {
		t.Error("token number")
	}
	for i, x := range res {
		if i%2 == 0 && x.Type != tkKeyword {
			t.Error("KEYWORD", x)
		}
	}
}
func TestLexer2(t *testing.T) {
	text := "case casex do do3 func func_ var var+ "
	lex := &Lexer{text, Tests}
	res, err := lex.Scan()
	if err != nil {
		t.Error(err)
	}
	if len(res) != 18 {
		t.Error(err)
	}
	for i, x := range res {
		if i == 0 || i == 4 || i == 8 || i == 12 || i == 14 {
			if x.Type != tkKeyword {
				t.Error("KEYWORD")
			}
		} else if x.Type == tkKeyword {
			t.Error("Should NOT KEYWORD", x)
		}
	}
}

func TestDebug(t *testing.T) {
	casedir, _ := filepath.Abs(filepath.Dir("./cases/"))
	outdir, _ := filepath.Abs(filepath.Dir("./" + testdata + "/"))
	option := Option{}
	option.IsDebug = true
	GenFile(casedir+"/var.gohtml", outdir+"/_var.gohtml", option)
}

func TestGenerate(t *testing.T) {
	casedir, _ := filepath.Abs(filepath.Dir("./cases/"))
	sap := string(filepath.Separator)

	visit := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() { // regular file
			if strings.HasPrefix(filepath.Base(path), ".") {
				return nil
			}
			name := strings.Replace(path, ".gohtml", ".go", 1)
			cmp := strings.Replace(name, sap+"cases"+sap, sap+testdata+sap, -1)
			dirname := filepath.Dir(cmp)
			log := filepath.Join(dirname, "_"+filepath.Base(cmp))
			if !exists(dirname) {
				os.MkdirAll(dirname, 0755)
			}
			option := Option{}

			if strings.HasSuffix(path, "panic.gohtml") {
				defer func() {
					if r := recover(); r != nil {
						panicMsg := fmt.Sprint(r)
						if !strings.HasPrefix(panicMsg, "failed to format template") ||
							!strings.Contains(panicMsg, ">>>> func Panic(totalMessage i nt) string {") {
							t.Error("panic.gohtml test failed")
						}
					}
				}()
				GenFile(path, log, option)
			} else {
				GenFile(path, log, option)
			}

			if !exists(cmp) {
				t.Error("No cmp:", cmp)
			} else if !exists(log) {
				t.Error("No log:", log)
			} else {
				//compare the log file and cmp file
				_cmp, _e1 := ioutil.ReadFile(cmp)
				_log, _e2 := ioutil.ReadFile(log)
				if _e1 != nil || _e2 != nil {
					t.Error("Reading")
				} else if string(_cmp) != string(_log) {
					t.Error("MISMATCH:", log, cmp)
				} else {
					t.Log("PASS")
				}
			}
		}
		return nil
	}
	QuickMode = true
	err := filepath.Walk(casedir, visit)
	if err != nil {
		t.Error("walk")
	}
}

func TestAdditionalCoverage(t *testing.T) {
	t.Run("exists_non_existent", func(t *testing.T) {
		if exists("non_existent_file_xyz_123") {
			t.Error("expected false")
		}
	})

	t.Run("getValStr_invalid_type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		getValStr(123)
	})

	t.Run("getValStr_ast_pointer", func(t *testing.T) {
		ast := &Ast{TagName: "my-tag"}
		if getValStr(ast) != "my-tag" {
			t.Error("expected my-tag")
		}
	})

	t.Run("ast_modestr_default", func(t *testing.T) {
		badAst := &Ast{Mode: 999}
		if badAst.ModeStr() != "UNDEF" {
			t.Error("expected UNDEF")
		}
	})

	t.Run("ast_check_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprint(r)
				if msg != "Maximum number of elements exceeded." {
					t.Error("unexpected panic:", msg)
				}
			} else {
				t.Error("expected panic")
			}
		}()
		panicAst := &Ast{}
		panicAst.Children = make([]interface{}, 100000)
		panicAst.check()
	})

	t.Run("ast_pop_empty", func(t *testing.T) {
		emptyAst := &Ast{}
		emptyAst.popChild() // should not panic
		if len(emptyAst.Children) != 0 {
			t.Error("expected 0 children")
		}
	})

	t.Run("ast_debug_and_has_non_exp", func(t *testing.T) {
		root := &Ast{Mode: PRG, TagName: "root"}
		child1 := &Ast{Mode: EXP, TagName: "exp_child"}
		child2 := &Ast{Mode: MKP, TagName: "markup_child"}
		root.addChild(child1)
		root.addChild(child2)
		root.debug(0, 5) // cover debug output
		root.debug(10, 5) // cover depth limit check

		if !root.hasNonExp() {
			t.Error("expected hasNonExp to be true")
		}
		if child1.hasNonExp() {
			t.Error("expected child1 (EXP with no children) hasNonExp to be false")
		}
	})

	t.Run("ast_has_non_exp_recursive_true", func(t *testing.T) {
		root := &Ast{Mode: EXP}
		child := &Ast{Mode: EXP}
		leaf := &Ast{Mode: MKP}
		child.addChild(leaf)
		root.addChild(child)
		if !root.hasNonExp() {
			t.Error("expected true")
		}
	})

	t.Run("isLayoutSection_direct", func(t *testing.T) {
		cp := &Compiler{
			isLayout:   true,
			paramNames: []string{"sidebar", "js"},
		}
		// test falls through on isLayoutSectionPart
		is, _ := cp.isLayoutSectionPart(Part{ptype: EXP, value: "not_a_write_string"})
		if is {
			t.Error("expected false")
		}

		// test match != in isLayoutSectionTest
		is, val := cp.isLayoutSectionTest(Part{ptype: CBLK, value: `if js != "" {`})
		if !is || val != "if js != nil {\n" {
			t.Error("expected true with match, got:", is, val)
		}
	})

	t.Run("settleLayout_prefix_adjust", func(t *testing.T) {
		defer func() {
			TemplateNamespacePrefix = ""
			recover() // we expect a panic because the layout file still doesn't exist
		}()
		TemplateNamespacePrefix = "myprefix"
		cp := &Compiler{
			layout: "myprefix/layout",
			file:   "somefile",
		}
		cp.settleLayout("NonExistentLayout")
	})

	t.Run("settleLayout_compilation_error", func(t *testing.T) {
		tmpLayout := filepath.Join(os.TempDir(), "temp_layout_error.gohtml")
		// Create a directory at tmpLayout so exists() is true but ReadFile fails
		err := os.MkdirAll(tmpLayout, 0755)
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpLayout)

		cp := &Compiler{
			layout: filepath.Dir(tmpLayout),
			file:   "somefile",
		}

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic due to compilation error in settleLayout")
			}
		}()

		cp.settleLayout(strings.TrimSuffix(filepath.Base(tmpLayout), ".gohtml"))
	})

	t.Run("run_non_existent", func(t *testing.T) {
		_, err := run("non_existent_file_xyz_123", Option{})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("parser_skip_token", func(t *testing.T) {
		parser := &Parser{
			tokens: []Token{
				{Type: tkContent, Text: "hello"},
				{Type: tkContent, Text: "world"},
			},
		}
		parser.skipToken()
		if parser.peekToken(0).Text != "world" {
			t.Error("expected next token to be world")
		}
	})

	t.Run("lexer_illegal_character", func(t *testing.T) {
		lex := &Lexer{"abc", []TokenMatch{}}
		_, err := lex.Scan()
		if err == nil || !strings.Contains(err.Error(), "Illegal character") {
			t.Error("expected Illegal character error, got:", err)
		}
	})

	t.Run("optimize_errors", func(t *testing.T) {
		// Parsing error
		ok, _ := optimize("dummy.go", "main", "package main\ninvalid syntax")
		if ok {
			t.Error("expected optimize to return false on syntax error")
		}

		// Type check error
		ok, _ = optimize("dummy.go", "main", "package main\nfunc main() { x := undefinedVar }")
		if ok {
			t.Error("expected optimize to return false on type checking error")
		}
	})

	t.Run("layout_args_zero_sections", func(t *testing.T) {
		SetLayout("tpl/layout/zero_args", []string{})
		cp := &Compiler{
			layout: "tpl/layout/zero_args",
		}
		foot := cp.generateFoot([]string{"mysection"})
		if !strings.Contains(foot, ", mysection()") {
			t.Error("expected section invocation in foot, got:", foot)
		}
	})

	t.Run("genfolder_errors", func(t *testing.T) {
		err := GenFolder("non_existent_folder_xyz_123", "out", Option{})
		if err == nil || !strings.Contains(err.Error(), "input directory does not exsit") {
			t.Error("expected folder not exist error, got:", err)
		}

		// GenFolder with non-existent outdir
		baseDir, _ := filepath.Abs(".")
		casedir := filepath.Join(baseDir, "cases")
		tmpOut := filepath.Join(os.TempDir(), "gorazor_test_out_dir_xyz_123")
		defer os.RemoveAll(tmpOut)

		// This should cover outdir creation in GenFolder
		err = GenFolder(filepath.Join(casedir, "layout"), tmpOut, Option{})
		if err != nil {
			t.Error("expected no error in GenFolder, got:", err)
		}
	})

	t.Run("noline_option", func(t *testing.T) {
		tmpOut := filepath.Join(os.TempDir(), "gorazor_noline_out.go")
		defer os.Remove(tmpOut)

		tmpGohtml := filepath.Join(os.TempDir(), "temp_noline_tpl.gohtml")
		defer os.Remove(tmpGohtml)

		err := ioutil.WriteFile(tmpGohtml, []byte(`@{
	var totalMessage int
}
<div>Hello @totalMessage</div>
`), 0644)
		if err != nil {
			t.Fatal(err)
		}

		option := Option{NoLineNumber: true}
		err = GenFile(tmpGohtml, tmpOut, option)
		if err != nil {
			t.Error("expected no error, got:", err)
		}

		content, err := ioutil.ReadFile(tmpOut)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(content), "// [") {
			t.Error("expected no line hint comments in generated file, but found them")
		}
	})

	t.Run("layout_not_found_panic", func(t *testing.T) {
		tmpOut := filepath.Join(os.TempDir(), "gorazor_layout_panic_out.go")
		defer os.Remove(tmpOut)

		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprint(r)
				if !strings.Contains(msg, "Can't find layout:") {
					t.Error("unexpected panic:", msg)
				}
			} else {
				t.Error("expected panic")
			}
		}()

		// Create a temporary layout file that references a non-existent layout
		tmpGohtml := filepath.Join(os.TempDir(), "temp_layout_xyz_123.gohtml")
		defer os.Remove(tmpGohtml)

		err := ioutil.WriteFile(tmpGohtml, []byte(`@{
	import (
		"tpl/layout"
	)
	layout := layout.NonExistentLayoutNameXYZ
}
<div>Hello</div>
`), 0644)
		if err != nil {
			t.Fatal(err)
		}

		option := Option{}
		// This should panic due to missing layout
		GenFile(tmpGohtml, tmpOut, option)
	})

	t.Run("genfolder_comprehensive", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "gorazor_comprehensive_test_dir_xyz")
		defer os.RemoveAll(tmpDir)

		inDir := filepath.Join(tmpDir, "in")
		outDir := filepath.Join(tmpDir, "out")
		os.MkdirAll(inDir, 0755)

		// 1. Create a valid .gohtml file
		ioutil.WriteFile(filepath.Join(inDir, "valid.gohtml"), []byte("<div>Hello</div>"), 0644)
		// 2. Create a non-.gohtml file (to cover !strings.HasSuffix branch)
		ioutil.WriteFile(filepath.Join(inDir, "dummy.txt"), []byte("ignore me"), 0644)
		// 3. Create a Emacs lock file (to cover strings.HasPrefix(\".#\") branch)
		ioutil.WriteFile(filepath.Join(inDir, ".#lock.gohtml"), []byte("lock file"), 0644)

		// Call GenFolder on the newly created directories (this also covers outDir not existing, i.e. !exists(outDir) in GenFolder)
		err := GenFolder(inDir, outDir, Option{})
		if err != nil {
			t.Fatal("GenFolder failed:", err)
		}

		// Verify files generated
		if !exists(filepath.Join(outDir, "valid.go")) {
			t.Error("expected valid.go to be generated")
		}
		if exists(filepath.Join(outDir, "dummy.go")) || exists(filepath.Join(outDir, ".#lock.go")) {
			t.Error("expected only valid.go to be generated")
		}

		// 4. Cover GenFile's output directory creation
		nonExistentSubDir := filepath.Join(tmpDir, "non_existent_subdir", "output.go")
		// this will call GenFile and hit the !exists(outdir) branch inside GenFile!
		err = GenFile(filepath.Join(inDir, "valid.gohtml"), nonExistentSubDir, Option{})
		if err != nil {
			t.Fatal("GenFile failed to create outdir:", err)
		}
	})

	t.Run("parser_prev_token_nil", func(t *testing.T) {
		parser := &Parser{
			preTokens: []Token{},
		}
		if parser.prevToken(0) != nil {
			t.Error("expected nil")
		}
	})

	t.Run("regMatch_invalid_regex", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		regMatch("[invalid", "text")
	})

	t.Run("regMatch_no_match", func(t *testing.T) {
		res, err := regMatch("abc", "def")
		if err != nil || res != "" {
			t.Error("expected empty string and no error")
		}
	})

	t.Run("parser_advance_until_escaped_end", func(t *testing.T) {
		parser := &Parser{
			tokens: []Token{
				{Type: tkAt, Text: "@"},
				{Type: tkParenClose, Text: ")"},
				{Type: tkParenClose, Text: ")"},
			},
		}
		tok := Token{Type: tkParenOpen, Text: "("}
		res, err := parser.advanceUntil(tok, tkParenOpen, tkParenClose, -1, tkAt)
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 4 {
			t.Error("expected 4 tokens, got:", len(res))
		}
	})

	t.Run("parser_advance_until_unmatched", func(t *testing.T) {
		parser := &Parser{
			tokens: []Token{},
		}
		tok := Token{Type: tkParenOpen, Text: "("}
		_, err := parser.advanceUntil(tok, tkParenOpen, tkParenClose, -1, tkAt)
		if err == nil || !strings.Contains(err.Error(), "Unmatched tag close") {
			t.Error("expected unmatched tag close error, got:", err)
		}
	})

	t.Run("parser_subparse_unmatched", func(t *testing.T) {
		parser := &Parser{
			tokens: []Token{},
		}
		tok := Token{Type: tkParenOpen, Text: "("}
		err := parser.subParse(tok, EXP, true)
		if err == nil || !strings.Contains(err.Error(), "Unmatched tag close") {
			t.Error("expected error, got:", err)
		}
	})
}

