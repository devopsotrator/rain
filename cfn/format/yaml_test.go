package format_test

import (
	"testing"

	"github.com/aws-cloudformation/rain/cfn/format"
)

func TestYamlDirectValues(t *testing.T) {
	cases := []interface{}{
		1,
		"foo",
		[]interface{}{1, "foo"},
		map[string]interface{}{"foo": "bar", "baz": "quux"},
		float64(500000000),
		float64(500000000.98765),
	}

	expecteds := []string{
		"1",
		"foo",
		"- 1\n- foo",
		"baz: quux\n\nfoo: bar",
		"500000000",
		"500000000.98765",
	}

	for i, testCase := range cases {
		expected := expecteds[i]

		actual := format.Anything(testCase, format.Options{})

		if actual != expected {
			t.Errorf("from %T %v:\n%#v != %#v\n", testCase, testCase, actual, expected)
		}
	}
}

func TestCompactYaml(t *testing.T) {
	cases := []interface{}{
		map[string]interface{}{"foo": "bar", "baz": "quux"},
	}

	expecteds := []string{
		"baz: quux\nfoo: bar",
	}

	for i, testCase := range cases {
		expected := expecteds[i]

		actual := format.Anything(testCase, format.Options{Compact: true})

		if actual != expected {
			t.Errorf("from %T %v:\n%#v != %#v\n", testCase, testCase, actual, expected)
		}
	}
}

func TestYamlScalars(t *testing.T) {
	cases := []map[string]interface{}{
		{"foo": 1},
		{"foo": 1.0},
		{"foo": 1.234},
		{"foo": float64(500000000)},
		{"foo": "32"},
		{"foo": "032"},
		{"foo": "32.0"},
		{"foo": "500000000"},
		{"foo": "hello"},
		{"foo": true},
		{"foo": false},
	}

	expecteds := []string{
		"foo: 1",
		"foo: 1",
		"foo: 1.234",
		"foo: 500000000",
		"foo: \"32\"",
		"foo: \"032\"",
		"foo: \"32.0\"",
		"foo: \"500000000\"",
		"foo: hello",
		"foo: true",
		"foo: false",
	}

	for i, testCase := range cases {
		expected := expecteds[i]

		actual := format.Anything(testCase, format.Options{})

		if actual != expected {
			t.Errorf("from %T %v:\n%#v != %#v\n", testCase, testCase, actual, expected)
		}
	}
}

func TestYamlList(t *testing.T) {
	cases := []map[string]interface{}{
		{"foo": []interface{}{}},
		{"foo": []interface{}{1}},
		{"foo": []interface{}{
			1,
			"foo",
			true,
		}},
		{"foo": []interface{}{
			[]interface{}{
				"foo",
				"bar",
			},
			"baz",
		}},
		{"foo": []interface{}{
			[]interface{}{
				[]interface{}{
					"foo",
					"bar",
				},
				"baz",
			},
			"quux",
		}},
		{"foo": []interface{}{
			map[string]interface{}{
				"foo": "bar",
			},
			map[string]interface{}{
				"baz":  "quux",
				"mooz": "xyzzy",
			},
		}},
	}

	expecteds := []string{
		"foo: []",
		"foo:\n  - 1",
		"foo:\n  - 1\n  - foo\n  - true",
		"foo:\n  - - foo\n    - bar\n  - baz",
		"foo:\n  - - - foo\n      - bar\n    - baz\n  - quux",
		"foo:\n  - foo: bar\n  - baz: quux\n    mooz: xyzzy",
	}

	for i, testCase := range cases {
		expected := expecteds[i]

		actual := format.Anything(testCase, format.Options{})

		if actual != expected {
			t.Errorf("from %T %v:\n%#v != %#v\n", testCase, testCase, actual, expected)
		}
	}
}

func TestYamlMap(t *testing.T) {
	cases := []map[string]interface{}{
		{},
		{
			"foo": "bar",
		},
		{
			"foo": "bar",
			"baz": "quux",
		},
		{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
			"quux": "mooz",
		},
		{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "quux",
				},
				"mooz": "xyzzy",
			},
			"alpha": "beta",
		},
		{
			"foo": []interface{}{
				"bar",
				"baz",
			},
			"quux": []interface{}{
				"mooz",
			},
		},
		{
			"32": map[string]interface{}{
				"64": "Numeric key",
			},
		},
	}

	expecteds := []string{
		"{}",
		"foo: bar",
		"baz: quux\n\nfoo: bar",
		"foo:\n  bar: baz\n\nquux: mooz",
		"alpha: beta\n\nfoo:\n  bar:\n    baz: quux\n\n  mooz: xyzzy",
		"foo:\n  - bar\n  - baz\n\nquux:\n  - mooz",
		"\"32\":\n  \"64\": Numeric key",
	}

	for i, testCase := range cases {
		expected := expecteds[i]

		actual := format.Anything(testCase, format.Options{})

		if actual != expected {
			t.Errorf("from %T %v:\n%#v != %#v\n", testCase, testCase, actual, expected)
		}
	}
}

func TestCfnYaml(t *testing.T) {
	cases := []map[string]interface{}{
		{
			"Quux":       "mooz",
			"Parameters": "baz",
			"Foo":        "bar",
			"Resources":  "xyzzy",
		},
	}

	expecteds := []string{
		"Parameters: baz\n\nResources: xyzzy\n\nFoo: bar\n\nQuux: mooz",
	}

	for i, testCase := range cases {
		expected := expecteds[i]

		actual := format.Anything(testCase, format.Options{})

		if actual != expected {
			t.Errorf("from %T %v:\n%#v != %#v\n", testCase, testCase, actual, expected)
		}
	}
}

func TestIntrinsics(t *testing.T) {
	cases := []map[string]interface{}{
		{
			"foo": map[string]interface{}{
				"Ref": "bar",
			},
		},
		{
			"foo": map[string]interface{}{
				"Fn::Sub": []interface{}{
					"The ${key} is a ${value}",
					map[string]interface{}{
						"key":   "cake",
						"value": "lie",
					},
				},
			},
		},
		{
			"foo": map[string]interface{}{
				"Fn::GetAtt": []interface{}{
					"Cake",
					"Lie",
				},
			},
		},
	}

	expecteds := []string{
		"foo: !Ref bar",
		"foo: !Sub\n  - The ${key} is a ${value}\n  - key: cake\n    value: lie",
		"foo: !GetAtt Cake.Lie",
	}

	for i, testCase := range cases {
		expected := expecteds[i]

		actual := format.Anything(testCase, format.Options{})

		if actual != expected {
			t.Errorf("from %T %v:\n%#v != %#v\n", testCase, testCase, actual, expected)
		}
	}
}

func TestStrings(t *testing.T) {
	cases := []map[string]interface{}{
		{"foo": "foo"},
		{"foo": "*"},
		{"foo": "* bar"},
		{"foo": "2012-05-02"},
		{"foo": "today is 2012-05-02"},
		{"foo": ": thing"},
		{"foo": "Yes"},
		{"foo": "No"},
		{"foo": "multi\nline"},
	}

	expecteds := []string{
		"foo: foo",
		"foo: \"*\"",
		"foo: \"* bar\"",
		"foo: \"2012-05-02\"",
		"foo: today is 2012-05-02",
		"foo: \": thing\"",
		"foo: \"Yes\"",
		"foo: \"No\"",
		"foo: |\n  multi\n  line",
	}

	for i, testCase := range cases {
		expected := expecteds[i]

		actual := format.Anything(testCase, format.Options{})

		if actual != expected {
			t.Errorf("from %T %v:\n%#v != %#v\n", testCase, testCase, actual, expected)
		}
	}
}

func TestYamlComments(t *testing.T) {
	data := map[string]interface{}{
		"foo": "bar",
		"baz": map[string]interface{}{
			"quux": "mooz",
		},
		"xyzzy": []interface{}{
			"lorem",
		},
	}

	commentCases := []map[interface{}]interface{}{
		{},
		{"": "Top-level comments"},
		{"foo": "This is foo"},
		{"baz": "This is baz"},
		{"baz": map[string]interface{}{"": "This is also baz"}},
		{"baz": map[string]interface{}{"quux": "This is quux"}},
		{"xyzzy": "This is xyzzy"},
		{"xyzzy": map[string]interface{}{"": "This is also xyzzy"}},
		{"xyzzy": map[interface{}]interface{}{0: "This is lorem"}}, // BUGGGGGGG
	}

	expecteds := []string{
		"baz:\n  quux: mooz\n\nfoo: bar\n\nxyzzy:\n  - lorem",
		"# Top-level comments\nbaz:\n  quux: mooz\n\nfoo: bar\n\nxyzzy:\n  - lorem",
		"baz:\n  quux: mooz\n\nfoo: bar  # This is foo\n\nxyzzy:\n  - lorem",
		"baz:  # This is baz\n  quux: mooz\n\nfoo: bar\n\nxyzzy:\n  - lorem",
		"baz:  # This is also baz\n  quux: mooz\n\nfoo: bar\n\nxyzzy:\n  - lorem",
		"baz:\n  quux: mooz  # This is quux\n\nfoo: bar\n\nxyzzy:\n  - lorem",
		"baz:\n  quux: mooz\n\nfoo: bar\n\nxyzzy:  # This is xyzzy\n  - lorem",
		"baz:\n  quux: mooz\n\nfoo: bar\n\nxyzzy:  # This is also xyzzy\n  - lorem",
		"baz:\n  quux: mooz\n\nfoo: bar\n\nxyzzy:\n  - lorem  # This is lorem",
	}

	for i, comments := range commentCases {
		expected := expecteds[i]

		actual := format.Anything(data, format.Options{
			Comments: comments,
		})

		if actual != expected {
			t.Errorf("from %q != %q\n", actual, expected)
		}
	}
}
