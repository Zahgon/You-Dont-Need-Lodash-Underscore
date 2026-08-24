package kebabcase

import "testing"

// ruleKeyCases covers every method key the plugin derives a rule name from,
// together with the exact output kebab-case@1.0.0 produces for it.
var ruleKeyCases = []struct {
	key  string
	want string
}{
	{"after", "after"},
	{"assign", "assign"},
	{"assignIn", "assign-in"},
	{"before", "before"},
	{"bind", "bind"},
	{"chunk", "chunk"},
	{"clamp", "clamp"},
	{"clone", "clone"},
	{"cloneDeep", "clone-deep"},
	{"compact", "compact"},
	{"concat", "concat"},
	{"defaults", "defaults"},
	{"defaultsDeep", "defaults-deep"},
	{"delay", "delay"},
	{"difference", "difference"},
	{"drop", "drop"},
	{"dropRight", "drop-right"},
	{"each", "each"},
	{"endsWith", "ends-with"},
	{"every", "every"},
	{"extend", "extend"},
	{"fill", "fill"},
	{"filter", "filter"},
	{"find", "find"},
	{"findIndex", "find-index"},
	{"findLast", "find-last"},
	{"findLastIndex", "find-last-index"},
	{"first", "first"},
	{"flatten", "flatten"},
	{"flattenDeep", "flatten-deep"},
	{"forEach", "for-each"},
	{"fromPairs", "from-pairs"},
	{"get", "get"},
	{"head", "head"},
	{"includes", "includes"},
	{"indexOf", "index-of"},
	{"isArray", "is-array"},
	{"isFunction", "is-function"},
	{"isInteger", "is-integer"},
	{"isNaN", "is-na-n"},
	{"isNil", "is-nil"},
	{"isNull", "is-null"},
	{"isString", "is-string"},
	{"isUndefined", "is-undefined"},
	{"join", "join"},
	{"keys", "keys"},
	{"last", "last"},
	{"lastIndexOf", "last-index-of"},
	{"map", "map"},
	{"omit", "omit"},
	{"padEnd", "pad-end"},
	{"padStart", "pad-start"},
	{"pick", "pick"},
	{"pluck", "pluck"},
	{"reduce", "reduce"},
	{"reduceRight", "reduce-right"},
	{"reject", "reject"},
	{"repeat", "repeat"},
	{"replace", "replace"},
	{"rest", "rest"},
	{"reverse", "reverse"},
	{"size", "size"},
	{"slice", "slice"},
	{"some", "some"},
	{"split", "split"},
	{"startsWith", "starts-with"},
	{"tail", "tail"},
	{"take", "take"},
	{"toLower", "to-lower"},
	{"toUpper", "to-upper"},
	{"trim", "trim"},
	{"values", "values"},
	{"capitalize", "capitalize"},
	{"at", "at"},
}

func TestKebabCaseRuleKeys(t *testing.T) {
	for _, testCase := range ruleKeyCases {
		t.Run(testCase.key, func(t *testing.T) {
			if got := KebabCase(testCase.key); got != testCase.want {
				t.Errorf("KebabCase(%q) = %q, want %q", testCase.key, got, testCase.want)
			}
		})
	}
}

// TestKebabCaseIsNaN pins the deliberately naive acronym handling that the
// generated rule names depend on.
func TestKebabCaseIsNaN(t *testing.T) {
	if got := KebabCase("isNaN"); got != "is-na-n" {
		t.Fatalf("KebabCase(\"isNaN\") = %q, want %q", got, "is-na-n")
	}
}

func TestKebabCaseEdgeCases(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"single lowercase", "a", "a"},
		{"single uppercase", "A", "-a"},
		{"leading uppercase", "Abc", "-abc"},
		{"all uppercase", "ABC", "-a-b-c"},
		{"digits untouched", "base64Encode", "base64-encode"},
		{"digits only", "12345", "12345"},
		{"hyphen untouched", "already-kebab", "already-kebab"},
		{"underscore untouched", "snake_Case", "snake_-case"},
		{"space untouched", "two Words", "two -words"},
		{"trailing uppercase", "endX", "end-x"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := KebabCase(testCase.in); got != testCase.want {
				t.Errorf("KebabCase(%q) = %q, want %q", testCase.in, got, testCase.want)
			}
		})
	}
}

func TestKebabCaseLatin1(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"U+00C0 A grave", "a\u00C0b", "a-\u00E0b"},
		{"U+00C6 AE", "a\u00C6b", "a-\u00E6b"},
		{"U+00D6 O diaeresis", "a\u00D6b", "a-\u00F6b"},
		{"U+00D7 multiplication sign excluded", "a\u00D7b", "a\u00D7b"},
		{"U+00D8 O stroke", "a\u00D8b", "a-\u00F8b"},
		{"U+00DE thorn", "a\u00DEb", "a-\u00FEb"},
		{"U+00DF sharp s excluded", "a\u00DFb", "a\u00DFb"},
		{"U+00BF inverted question excluded", "a\u00BFb", "a\u00BFb"},
		{"lowercase latin1 excluded", "a\u00E0b", "a\u00E0b"},
		{"non latin greek excluded", "a\u0391b", "a\u0391b"},
		{"non latin cyrillic excluded", "a\u0410b", "a\u0410b"},
		{"cjk excluded", "a\u4E00b", "a\u4E00b"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := KebabCase(testCase.in); got != testCase.want {
				t.Errorf("KebabCase(%q) = %q, want %q", testCase.in, got, testCase.want)
			}
		})
	}
}

func TestIsKebabTriggerBoundaries(t *testing.T) {
	cases := []struct {
		name string
		r    rune
		want bool
	}{
		{"before A", '@', false},
		{"A", 'A', true},
		{"Z", 'Z', true},
		{"after Z", '[', false},
		{"lowercase a", 'a', false},
		{"digit", '5', false},
		{"below C0", 0x00BF, false},
		{"C0", 0x00C0, true},
		{"D6", 0x00D6, true},
		{"D7", 0x00D7, false},
		{"D8", 0x00D8, true},
		{"DE", 0x00DE, true},
		{"DF", 0x00DF, false},
		{"far above range", 0x1F600, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := isKebabTrigger(testCase.r); got != testCase.want {
				t.Errorf("isKebabTrigger(%U) = %v, want %v", testCase.r, got, testCase.want)
			}
		})
	}
}
