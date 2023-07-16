package new_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.tesla.cn/itapp/lines/logx"
	"github.com/go_service/internal/pen/new"
	"github.com/go_service/internal/pen/pkg/options"
)

var _ = Describe("graceTerm Package", func() {

	BeforeSuite(func() {
		logx.Init(true)
	})

	It("test generate module", func() {
		const toCompare = "tocompare"
		opts := options.Options{
			ModuleName: "testmodule",
			AppPackage: "github.com/go_service/internal/pen/new/testmodule",
		}
		err := new.GenerateNewModules(opts)
		Expect(err).To(BeNil())
		for _, fileName := range []string{"Makefile", opts.ModuleName + ".yaml", "pen.yaml"} {
			ExpectSameFile(path.Join(opts.ModuleName, fileName), path.Join(toCompare, fileName))
		}

		err = os.RemoveAll(opts.ModuleName)
		Expect(err).To(BeNil())
	})
})

func ExpectSameFile(a, b string) {
	fa, err := ioutil.ReadFile(a)
	Expect(err).To(BeNil())
	fb, err := ioutil.ReadFile(b)
	Expect(err).To(BeNil())
	Expect(string(fa)).To(Equal(string(fb)))
}

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Pen Test Suite", []Reporter{junitReporter})
}
