package log

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

type testTextMarshal struct{}

func (t testTextMarshal) MarshalText() ([]byte, error) {
	return []byte(`a", b, c`), nil
}

type testJSONMarshal struct{}

func (t testJSONMarshal) MarshalJSON() ([]byte, error) {
	return []byte(`"a\", b, c"` + " \r\n "), nil
}

type testError struct{}

func (t testError) Error() string {
	return `a", b, c`
}

func TestJSONFormat(t *testing.T) {
	t.Parallel()

	l := NewLogger()
	l.SetTopic("topic1")
	l.SetDefaults(map[string]interface{}{
		"abc":   123,
		"_d123": true,
	})

	ts := time.Date(2001, time.December, 3, 13, 45, 1, 123456789, time.UTC)
	f := JSONFormat{"localhost"}
	buf := make([]byte, 0, 4096)

	b, err := f.Format(buf, l, ts, LvCritical, "hoge", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 || b[len(b)-1] != '\n' {
		t.Error(`len(b) == 0 || b[len(b)-1] != '\n'`)
	}
	var j map[string]interface{}
	err = json.Unmarshal(b, &j)
	if err != nil {
		t.Fatal(err)
	}

	if v, ok := j[FnTopic]; !ok {
		t.Error(`v, ok := j[FnTopic]; !ok`)
	} else {
		if v.(string) != "topic1" {
			t.Error(`v.(string) != "topic1"`)
		}
	}

	if v, ok := j[FnLoggedAt]; !ok {
		t.Error(`v, ok := j[FnLoggedAt]; !ok`)
	} else {
		if v.(string) != "2001-12-03T13:45:01.123456Z" {
			t.Error(`v.(string) != "2001-12-03T13:45:01.123456Z"`)
		}
	}

	if v, ok := j[FnUtsname]; !ok {
		t.Error(`v, ok := j[FnUtsname]; !ok`)
	} else {
		if v.(string) != "localhost" {
			t.Error(`v.(string) != "localhost"`)
		}
	}

	if v, ok := j[FnSeverity]; !ok {
		t.Error(`v, ok := j[FnSeverity]; !ok`)
	} else {
		if v.(string) != "critical" {
			t.Error(`v.(string) != "critical"`)
		}
	}

	if v, ok := j[FnMessage]; !ok {
		t.Error(`v, ok := j[FnMessage]; !ok`)
	} else {
		if v.(string) != "hoge" {
			t.Error(`v.(string) != "hoge"`)
		}
	}

	if v, ok := j["abc"]; !ok {
		t.Error(`v, ok := j["abc"]; !ok`)
	} else {
		if int(v.(float64)) != 123 {
			t.Error(`int(v.(float64)) != 123`)
		}
	}

	if v, ok := j["_d123"]; !ok {
		t.Error(`v, ok := j["_d123"]; !ok`)
	} else {
		if !v.(bool) {
			t.Error(`!v.(bool)`)
		}
	}

	b, err = f.Format(buf, l, ts, LvDebug, "fuga fuga", map[string]interface{}{
		"abc":     []int{1, 2, 3},
		"invalid": "12\xc534",
		"tm":      testTextMarshal{},
		"jm":      testJSONMarshal{},
		"err":     testError{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 || b[len(b)-1] != '\n' {
		t.Error(`len(b) == 0 || b[len(b)-1] != '\n'`)
	}
	// JSON Lines must not contain \n
	if bytes.ContainsAny(b[:len(b)-1], "\r\n") {
		t.Error(`bytes.ContainsAny(b[:len(b)-1], "\r\n")`)
	}
	j = nil
	err = json.Unmarshal(b, &j)
	if err != nil {
		t.Fatal(err)
	}

	if v, ok := j[FnSeverity]; !ok {
		t.Error(`v, ok := j[FnSeverity]; !ok`)
	} else {
		if v.(string) != "debug" {
			t.Error(`v.(string) != "debug"`)
		}
	}

	if v, ok := j[FnMessage]; !ok {
		t.Error(`v, ok := j[FnMessage]; !ok`)
	} else {
		if v.(string) != "fuga fuga" {
			t.Error(`v.(string) != "fuga fuga"`)
		}
	}

	if v, ok := j["abc"]; !ok {
		t.Error(`v, ok := j["abc"]; !ok`)
	} else {
		if !reflect.DeepEqual(v.([]interface{}), []interface{}{1.0, 2.0, 3.0}) {
			t.Error(`!reflect.DeepEqual(v.([]interface{}), []interface{}{1, 2, 3})`)
			t.Logf("%#v", v.([]interface{}))
		}
	}

	if v, ok := j["_d123"]; !ok {
		t.Error(`v, ok := j["_d123"]; !ok`)
	} else {
		if !v.(bool) {
			t.Error(`!v.(bool)`)
		}
	}

	if v, ok := j["invalid"]; !ok {
		t.Error(`v, ok := j["invalid"]; !ok`)
	} else {
		if v.(string) != "12\uFFFD34" {
			t.Error(`v.(string) != "12\uFFFD34"`)
		}
	}

	if v, ok := j["tm"]; !ok {
		t.Error(`v, ok := j["tm"]; !ok`)
	} else {
		if v.(string) != `a", b, c` {
			t.Error("v.(string) != `a\", b, c`")
		}
	}

	if v, ok := j["jm"]; !ok {
		t.Error(`v, ok := j["jm"]; !ok`)
	} else {
		if v.(string) != `a", b, c` {
			t.Error("v.(string) != `a\", b, c`")
		}
	}

	if v, ok := j["err"]; !ok {
		t.Error(`v, ok := j["err"]; !ok`)
	} else {
		if v.(string) != `a", b, c` {
			t.Error("v.(string) != `a\", b, c`")
		}
	}
}
