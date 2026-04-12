package asr

import "testing"

func TestSimplifyZh_ConvertsTraditionalCharacters(t *testing.T) {
	got := SimplifyZh("歡迎來到這裡說中文")
	want := "欢迎来到这里说中文"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSimplifyZh_LeavesUnsupportedCharactersUnchanged(t *testing.T) {
	got := SimplifyZh("ABC 123 already simple")
	want := "ABC 123 already simple"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
