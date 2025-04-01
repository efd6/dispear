package dispear

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/gotooltest"
	"github.com/rogpeppe/go-internal/testscript"
)

var (
	update = flag.Bool("update", false, "update tests")
	keep   = flag.Bool("keep", false, "keep $WORK directory after tests")
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, nil))
}

func TestScripts(t *testing.T) {
	t.Parallel()

	p := testscript.Params{
		Dir:           filepath.Join("testdata"),
		UpdateScripts: *update,
		TestWork:      *keep,
		Setup: func(e *testscript.Env) error {
			pwd, err := os.Getwd()
			if err != nil {
				return err
			}
			e.Setenv("PKG_ROOT", pwd)
			return nil
		},
	}
	if err := gotooltest.Setup(&p); err != nil {
		t.Fatal(err)
	}
	testscript.Run(t, p)
}
