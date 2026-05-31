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
		baseDir, _ := filepath.Abs(".")
		casedir := filepath.Join(baseDir, "cases")
		tmpOut := filepath.Join(os.TempDir(), "gorazor_noline_out.go")
		defer os.Remove(tmpOut)

		option := Option{NoLineNumber: true}
		err := GenFile(filepath.Join(casedir, "var.gohtml"), tmpOut, option)
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
}

