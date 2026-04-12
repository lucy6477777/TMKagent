package asr

import "testing"

func TestSimplifyZh_TraditionalToSimplified(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"說話", "说话"},
		{"這個國家", "这个国家"},
		{"電視機", "电视机"},
		{"謝謝你們", "谢谢你们"},
	}
	for _, c := range cases {
		got := SimplifyZh(c.in)
		if got != c.want {
			t.Errorf("SimplifyZh(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSimplifyZh_AlreadySimplified(t *testing.T) {
	input := "你好世界"
	got := SimplifyZh(input)
	if got != input {
		t.Errorf("SimplifyZh(%q) = %q, want unchanged", input, got)
	}
}

func TestSimplifyZh_MixedContent(t *testing.T) {
	got := SimplifyZh("Hello 說 world 们")
	want := "Hello 说 world 们"
	if got != want {
		t.Errorf("SimplifyZh mixed = %q, want %q", got, want)
	}
}

func TestSimplifyZh_Empty(t *testing.T) {
	got := SimplifyZh("")
	if got != "" {
		t.Errorf("SimplifyZh empty = %q, want empty", got)
	}
}

func TestSimplifyZh_NoChineseCharacters(t *testing.T) {
	input := "Hello World 123"
	got := SimplifyZh(input)
	if got != input {
		t.Errorf("SimplifyZh(%q) = %q, want unchanged", input, got)
	}
}
