package fs

import (
	"context"
	"github.com/spf13/afero"
	"github.com/phyrwork/mobius/filter"
	"os"
	"github.com/phyrwork/mobius/fs"
	"fmt"
)

type Importer struct {
	io afero.Fs
	Filter filter.Filter
}

func NewImporter(io afero.Fs, root string) Importer {
	io = afero.NewBasePathFs(io, root)
	return Importer{io: io}
}

func (im Importer) Import(ctx context.Context, dst *fs.Fs) error {
	return afero.Walk(im.io, "", func (path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Ignore root directory
		if path == "" {
			return nil
		}
		// Filter
		if im.Filter != nil {
			pass, err := im.Filter.Filter(path)
			if err != nil {
				return err
			}
			if !pass {
				return nil
			}
		}
		if _, err := dst.Create(ctx, path); err != nil {
			return fmt.Errorf("error creating graph file %v: %v", path, err)
		}
		return nil
	})
}