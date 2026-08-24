# Untranslatable cases from `tests/unit/all.js`

`tests_unit_all.js` is the original JavaScript suite, carried across byte for
byte as evidence. It is not executed by anything in this repository; it is here
so the exclusions below can be checked against the real source.

That suite does not test any artifact of this repository. It requires `lodash`
and `assert` and asserts that each lodash function agrees with the hand-written
native JavaScript snippet documented in the README. Its unit under test is
therefore the JavaScript language plus the lodash library.

89 of its 128 `it()` cases are translated in `tests/unit/*_test.go`, with the
helper pairs they exercise in `internal/lodashdoc/`. The remaining 39 are
excluded by **impossibility, not inconvenience**: each one is true only because
of a JavaScript value-model feature that Go does not have (`undefined` as a
value, ToBoolean coercion across unrelated types, the prototype chain,
primitive-versus-wrapper-object identity, or lodash's chainable `_()` wrapper).
Porting them would have meant building a JavaScript interpreter's value model in
Go, which would test the emulation rather than the behaviour.

See `manifest.json` for the machine-readable list: every excluded case with its
describe title, its verbatim `it()` title, its source line, its reason code, and
a one-sentence reason.
