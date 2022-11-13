//go:build go1.19

package html2text

import "testing"

func FuzzFromString(f *testing.F) {
	f.Add("Hello, world!")
	f.Add("Hello, 世界!")
	f.Add("Hello, <b>world!</b>")
	f.Add("Hello, <b>世界!</b>")
	f.Add("<p>Hello, world!</p>")
	f.Add("<p>Hello, 世界!</p>")
	f.Add("<p>Hello, <b>world!</b></p>")
	f.Add("<p>こんにちは</p>")
	f.Fuzz(func(t *testing.T, s string) {
		text, err := FromString(s)
		if err != nil && text != "" {
			t.Errorf("%q, %v", text, err)
		}
	})
}
