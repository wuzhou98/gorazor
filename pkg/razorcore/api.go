package razorcore

//------------------------------ API ------------------------------

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	goExtension = ".go"
	gzExtension = ".gohtml"
)

// GenFile generate from input to output file,
// gofmt will trigger an error if it fails.
func GenFile(input string, output string, options Option) error {
	outdir := filepath.Dir(output)
	if !exists(outdir) {
		os.MkdirAll(outdir, 0775)
	}
	if options.LayoutCache == nil {
		options.LayoutCache = NewLayoutCache()
	}
	return generate(input, output, options)
}

// GenFolder generate from directory to directory, Find all the files with extension
// of .gohtml and generate it into target dir.
func GenFolder(indir string, outdir string, options Option) (err error) {
	if !exists(indir) {
		return errors.New("input directory does not exsit")
	}

	if options.LayoutCache == nil {
		options.LayoutCache = NewLayoutCache()
	}

	//Make it
	if !exists(outdir) {
		os.MkdirAll(outdir, 0775)
	}

	incdirAbs, _ := filepath.Abs(indir)
	outdirAbs, _ := filepath.Abs(outdir)

	paths := []string{}

	visit := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			//Just do file with exstension .gohtml
			if !strings.HasSuffix(path, gzExtension) {
				return nil
			}
			filename := filepath.Base(path)
			if strings.HasPrefix(filename, ".#") {
				return nil
			}
			paths = append(paths, path)
		}
		return nil
	}

	var errMutex sync.Mutex
	var firstErr error

	fun := func(path string, res chan<- string) {
		//adjust with the abs path, so that we keep the same directory hierarchy
		input, _ := filepath.Abs(path)
		output := strings.Replace(input, incdirAbs, outdirAbs, 1)
		output = strings.ReplaceAll(output, gzExtension, goExtension)
		genErr := GenFile(path, output, options)
		if genErr != nil {
			errMutex.Lock()
			if firstErr == nil {
				firstErr = genErr
			}
			errMutex.Unlock()
		}
		res <- fmt.Sprintf("%s -> %s", path, output)
	}

	err = filepath.Walk(indir, visit)
	if err != nil {
		return err
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	result := make(chan string, len(paths))

	for w := 0; w < len(paths); w++ {
		go fun(paths[w], result)
	}
	for i := 0; i < len(paths); i++ {
		res := <-result
		fmt.Println(res)
	}

	return firstErr
}
