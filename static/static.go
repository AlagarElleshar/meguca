// Static file storage embedded into the binary

//go:generate statik --src ./src -f

package static

import (
	"github.com/go-playground/log"
	"net/http"
	"os"

	_ "github.com/bakape/meguca/static/statik"
	"github.com/rakyll/statik/fs"
)

var (
	// Embedded in-binary filesystem. Contained files must not be modified.
	FS http.FileSystem
)

func init() {
	var err error
	FS, err = fs.New()
	if err != nil {
		panic(err)
	}
}

// Read file from embedded file system into buffer
func ReadFile(path string) (buf []byte, err error) {
	path2, err := os.Getwd()
	if err != nil {
		log.Info(err)
	}
	log.Info(path2)
	log.Info("Path: ", path)
	log.Info("Complete path: ", path2+path)
	return fs.ReadFile(FS, path)
}
