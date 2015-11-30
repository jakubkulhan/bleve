package hunspell

import (
	"testing"
	"github.com/blevesearch/bleve/analysis"
)

func TestHunspellFilter(t *testing.T) {

	tests := []struct {
		affFile string
		dicFiles []string
		input  analysis.TokenStream
		output analysis.TokenStream
	}{
		{
			input: analysis.TokenStream{
				&analysis.Token{
					Term: []byte("drinkable"),
				},
			},
			output: analysis.TokenStream{
				&analysis.Token{
					Term: []byte("drink"),
				},
			},
		},
	}

	_ = tests

//	for _, test := range tests {
//		hunspellFilter, err := HunspellFilterConstructor(map[string]interface{} {
//			"aff_file": test.affFile,
//			"dic_files": test.dicFiles,
//		}, nil)
//
//		if err != nil {
//			t.Error(err)
//			continue
//		}
//
//		actual := hunspellFilter.Filter(test.input)
//		if !reflect.DeepEqual(actual, test.output) {
//			t.Errorf("expected %s, got %s", test.output, actual)
//		}
//	}

}
