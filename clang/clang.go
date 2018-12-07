package clang

import (
	"github.com/phyrwork/mobius/fs"
)

type File struct {
	fs.File
	Includes []Include
}