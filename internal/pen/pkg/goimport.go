package pkg

import (
	"github.com/go_service/internal/importformat/pkg/importformat"
)

func ImportFormat(b string) string {
	b = importformat.GoImports(b)
	return importformat.FormatImport(b)
}
