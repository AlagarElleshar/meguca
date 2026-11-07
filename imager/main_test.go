package imager

import (
	"fmt"
	"os"
	"testing"

	"github.com/bakape/meguca/config"
	"github.com/bakape/meguca/db"
	"github.com/bakape/meguca/imager/assets"
	. "github.com/bakape/meguca/test"
)

func TestMain(m *testing.M) {
	fmt.Println("Entering TestMain")
	err := config.Server.Load()
	fmt.Println("Loading server config")
	if err != nil {
		panic(err)
	}
	close, err := db.LoadTestDB("imager")
	fmt.Println("Loading test DB")
	if err != nil {
		panic(err)
	}

	config.Set(config.Configs{})
	fmt.Println("Setting config")
	if err := assets.CreateDirs(); err != nil {
		panic(err)
	}

	code := m.Run()
	fmt.Println("Running code")
	err = close()
	fmt.Println("Close finishing")
	if err != nil {
		fmt.Println("Close panic")
		panic(err)
	}
	err = assets.DeleteDirs()
	fmt.Println("Deleting dirs")
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}

func resetDirs(t *testing.T) {
	t.Helper()
	if err := assets.ResetDirs(); err != nil {
		t.Fatal(err)
	}
}

// How do we assert a thumbnail?
func assertThumbnail(t *testing.T, thumb []byte) {
	t.Helper()
	if thumb != nil && len(thumb) < 100 {
		t.Fatal("thumbnail too small")
	}
}

func assertDims(t *testing.T, res, std [4]uint16) {
	t.Helper()
	if res != std {
		LogUnexpected(t, std, res)
	}
}
