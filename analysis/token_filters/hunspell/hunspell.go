package hunspell

import (
	"github.com/blevesearch/bleve/analysis"
	"github.com/blevesearch/bleve/registry"
	"fmt"
	"io/ioutil"
)

const (
	Name = "hunspell"
)

type HunspellFilter struct {

}

func NewHunspellFilter(affFile string, dicFiles []string) *HunspellFilter {
	affContent, err := ioutil.ReadFile(affFile)

	if err != nil {
		panic(err)
	}

	_ = affContent

	// TODO: dicFiles

	return &HunspellFilter{

	}
}

func (f *HunspellFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	return input
}

func HunspellFilterConstructor(config map[string]interface{}, cache *registry.Cache) (analysis.TokenFilter, error) {
	affFile, ok := config["aff_file"].(string)
	if !ok {
		return nil, fmt.Errorf("must specify aff_file")
	}
	dicFiles, ok := config["dic_files"].([]string)
	if !ok {
		dicFile, ok := config["dic_file"].(string)
		if !ok {
			return nil, fmt.Errorf("must specify dic_file(s)")
		}

		dicFiles = make([]string, 1)
		dicFiles[0] = dicFile
	}

	return NewHunspellFilter(affFile, dicFiles), nil
}

func init() {
	registry.RegisterTokenFilter(Name, HunspellFilterConstructor)
}
