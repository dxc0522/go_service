package new

import (
	"bufio"
	"bytes"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"

	"github.tesla.cn/itapp/lines/errorx"
	"github.tesla.cn/itapp/lines/filex"
	"github.tesla.cn/itapp/lines/logx"
	"github.com/go_service/internal/pen/pkg"
	"github.com/go_service/internal/pen/pkg/options"
)

type (
	GeneratorFunc func() error
	Generator     struct {
		*openapi3.T
		*template.Template
		options.Options
		TemplateName string
		TargetFile   string
		PenFile      bool // auto regenerated file
	}
)

func (g Generator) writeFile(target string, toBeFormat []byte) error {
	result := toBeFormat
	if strings.HasSuffix(target, ".go") {
		var err error
		content := []byte(pkg.ImportFormat(string(toBeFormat)))
		newContent, err := format.Source(content)
		if err != nil {
			return errorx.WithStack(err)
		}
		result = newContent
	}
	filex.EnsureDir(filepath.Dir(target))
	if exists, err := filex.Exists(target); err != nil {
		return err
	} else if exists {
		logx.Info("skip non pen file", "file", target)
		return nil
	}
	return errorx.WithStack(ioutil.WriteFile(target, result, 0644))
}

func (g Generator) ExecuteTemplate(context interface{}) ([]byte, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	err := g.Template.ExecuteTemplate(w, g.TemplateName+pkg.TemplateSuffix, context)
	if err != nil {
		return nil, errorx.Wrap(err, "error when generating "+g.TemplateName)
	}
	err = w.Flush()
	if err != nil {
		return nil, errorx.Wrap(err, "error when generating "+g.TemplateName)
	}
	return buf.Bytes(), nil
}
