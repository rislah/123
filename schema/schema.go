package schema

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"strings"

	"github.com/rislah/fakes/internal/errors"
)

//go:embed *.graphql
//go:embed */**.graphql
var schemaFiles embed.FS

func String() (string, error) {
	var buf bytes.Buffer

	fn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".graphql") {
			return nil
		}

		b, err := schemaFiles.ReadFile(path)
		if err != nil {
			return err
		}

		b = append(b, []byte("\n")...)

		if _, err := buf.Write(b); err != nil {
			return errors.Wrap(err, "writing bytes to buffer")
		}

		return nil
	}

	if err := fs.WalkDir(schemaFiles, ".", fn); err != nil {
		return buf.String(), errors.Wrap(err, "walking dir")
	}

	return buf.String(), nil
}
