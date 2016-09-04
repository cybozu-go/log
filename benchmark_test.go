package log

import (
	"io/ioutil"
	"testing"
)

func makeFields() map[string]interface{} {
	return map[string]interface{}{
		"str":   "abc def ghi",
		"int":   int(-12345),
		"float": float64(3.14159),
		"slice": []int{1, 2, 3},
	}
}

func BenchmarkPlain(b *testing.B) {
	l := NewLogger()
	l.SetOutput(ioutil.Discard)
	l.SetFormatter(PlainFormat{})
	fields := makeFields()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Error("test", fields)
	}
}

func BenchmarkLogfmt(b *testing.B) {
	l := NewLogger()
	l.SetOutput(ioutil.Discard)
	l.SetFormatter(Logfmt{})
	fields := makeFields()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Error("test", fields)
	}
}

func BenchmarkJSON(b *testing.B) {
	l := NewLogger()
	l.SetOutput(ioutil.Discard)
	l.SetFormatter(JSONFormat{})
	fields := makeFields()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Error("test", fields)
	}
}
