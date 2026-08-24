package estree

import "testing"

func TestMemberExpressions(t *testing.T) {
	t.Run("dot access is not computed", func(t *testing.T) {
		program := mustParse(t, "_.map(list);")
		members := collect(program, "MemberExpression")
		if len(members) != 1 {
			t.Fatalf("found %d MemberExpression nodes, want 1", len(members))
		}
		member := members[0].(*MemberExpression)
		if member.Computed {
			t.Errorf("Computed = true, want false")
		}
		if object, ok := member.Object.(*Identifier); !ok || object.Name != "_" {
			t.Errorf("Object = %v, want Identifier _", member.Object)
		}
		if property, ok := member.Property.(*Identifier); !ok || property.Name != "map" {
			t.Errorf("Property = %v, want Identifier map", member.Property)
		}
	})

	t.Run("bracket access with identifier index is computed", func(t *testing.T) {
		program := mustParse(t, "var v = target[key];")
		member := collect(program, "MemberExpression")[0].(*MemberExpression)
		if !member.Computed {
			t.Errorf("Computed = false, want true")
		}
		if property, ok := member.Property.(*Identifier); !ok || property.Name != "key" {
			t.Errorf("Property = %v, want Identifier key", member.Property)
		}
	})

	t.Run("bracket access with string index is computed", func(t *testing.T) {
		program := mustParse(t, "var v = target['name'];")
		member := collect(program, "MemberExpression")[0].(*MemberExpression)
		if !member.Computed {
			t.Errorf("Computed = false, want true")
		}
		if property, ok := member.Property.(*Literal); !ok || property.Value != "name" {
			t.Errorf("Property = %v, want Literal name", member.Property)
		}
	})
}

func TestObjectKeys(t *testing.T) {
	t.Run("bare key is an Identifier", func(t *testing.T) {
		program := mustParse(t, "var o = { alpha: 1 };")
		properties := collect(program, "Property")
		if len(properties) != 1 {
			t.Fatalf("found %d Property nodes, want 1", len(properties))
		}
		property := properties[0].(*Property)
		key, ok := property.Key.(*Identifier)
		if !ok {
			t.Fatalf("Key is %T, want *Identifier", property.Key)
		}
		if key.Name != "alpha" {
			t.Errorf("Key.Name = %q, want alpha", key.Name)
		}
	})

	t.Run("single quoted key is a Literal", func(t *testing.T) {
		program := mustParse(t, "var o = { 'alpha': 1 };")
		property := collect(program, "Property")[0].(*Property)
		key, ok := property.Key.(*Literal)
		if !ok {
			t.Fatalf("Key is %T, want *Literal", property.Key)
		}
		if key.Value != "alpha" || !key.IsString {
			t.Errorf("Key = %+v, want string Literal alpha", key)
		}
	})

	t.Run("double quoted key is a Literal", func(t *testing.T) {
		program := mustParse(t, "var o = { \"alpha\": 1 };")
		property := collect(program, "Property")[0].(*Property)
		if _, ok := property.Key.(*Literal); !ok {
			t.Fatalf("Key is %T, want *Literal", property.Key)
		}
	})

	t.Run("shorthand key is marked", func(t *testing.T) {
		program := mustParse(t, "var alpha = 1; var o = { alpha };")
		property := collect(program, "Property")[0].(*Property)
		if !property.Shorthand {
			t.Errorf("Shorthand = false, want true")
		}
	})

	t.Run("non shorthand key is not marked", func(t *testing.T) {
		program := mustParse(t, "var beta = 1; var o = { alpha: beta };")
		property := collect(program, "Property")[0].(*Property)
		if property.Shorthand {
			t.Errorf("Shorthand = true, want false")
		}
	})

	t.Run("object spread becomes a RestElement", func(t *testing.T) {
		program := mustParse(t, "var o = { ...other };")
		rests := collect(program, "RestElement")
		if len(rests) != 1 {
			t.Fatalf("found %d RestElement nodes, want 1", len(rests))
		}
		rest := rests[0].(*RestElement)
		if argument, ok := rest.Argument.(*Identifier); !ok || argument.Name != "other" {
			t.Errorf("Argument = %v, want Identifier other", rest.Argument)
		}
	})
}

func TestDestructuring(t *testing.T) {
	t.Run("object pattern binding", func(t *testing.T) {
		program := mustParse(t, "var { map, filter } = require('lodash');")
		patterns := collect(program, "ObjectPattern")
		if len(patterns) != 1 {
			t.Fatalf("found %d ObjectPattern nodes, want 1", len(patterns))
		}
		pattern := patterns[0].(*ObjectPattern)
		if len(pattern.Properties) != 2 {
			t.Fatalf("Properties has %d entries, want 2", len(pattern.Properties))
		}
		first := pattern.Properties[0].(*Property)
		if key := first.Key.(*Identifier); key.Name != "map" {
			t.Errorf("first Key = %q, want map", key.Name)
		}
		if !first.Shorthand {
			t.Errorf("first Shorthand = false, want true")
		}
	})

	t.Run("renamed object pattern binding", func(t *testing.T) {
		program := mustParse(t, "var { map: mapper } = require('lodash');")
		pattern := collect(program, "ObjectPattern")[0].(*ObjectPattern)
		property := pattern.Properties[0].(*Property)
		if key := property.Key.(*Identifier); key.Name != "map" {
			t.Errorf("Key = %q, want map", key.Name)
		}
		if value := property.Value.(*Identifier); value.Name != "mapper" {
			t.Errorf("Value = %q, want mapper", value.Name)
		}
		if property.Shorthand {
			t.Errorf("Shorthand = true, want false")
		}
	})

	t.Run("quoted key in object pattern", func(t *testing.T) {
		program := mustParse(t, "var { 'map': mapper } = require('lodash');")
		pattern := collect(program, "ObjectPattern")[0].(*ObjectPattern)
		property := pattern.Properties[0].(*Property)
		if _, ok := property.Key.(*Literal); !ok {
			t.Fatalf("Key is %T, want *Literal", property.Key)
		}
	})

	t.Run("rest binding in object pattern", func(t *testing.T) {
		program := mustParse(t, "var { map, ...others } = require('lodash');")
		pattern := collect(program, "ObjectPattern")[0].(*ObjectPattern)
		if len(pattern.Properties) != 2 {
			t.Fatalf("Properties has %d entries, want 2", len(pattern.Properties))
		}
		rest, ok := pattern.Properties[1].(*RestElement)
		if !ok {
			t.Fatalf("Properties[1] is %T, want *RestElement", pattern.Properties[1])
		}
		if argument, ok := rest.Argument.(*Identifier); !ok || argument.Name != "others" {
			t.Errorf("Argument = %v, want Identifier others", rest.Argument)
		}
	})

	t.Run("array pattern binding becomes Unknown", func(t *testing.T) {
		program := mustParse(t, "var [first, second] = list;")
		unknowns := collect(program, "Unknown")
		if len(unknowns) == 0 {
			t.Fatalf("found no Unknown node for the array pattern")
		}
		if kind := unknowns[0].(*Unknown).Kind; kind != "BArray" {
			t.Errorf("Kind = %q, want BArray", kind)
		}
	})
}

func TestAssignmentExpressions(t *testing.T) {
	t.Run("identifier target", func(t *testing.T) {
		program := mustParse(t, "var target; target = source;")
		assignments := collect(program, "AssignmentExpression")
		if len(assignments) != 1 {
			t.Fatalf("found %d AssignmentExpression nodes, want 1", len(assignments))
		}
		assignment := assignments[0].(*AssignmentExpression)
		if assignment.Operator != "=" {
			t.Errorf("Operator = %q, want =", assignment.Operator)
		}
		if left, ok := assignment.Left.(*Identifier); !ok || left.Name != "target" {
			t.Errorf("Left = %v, want Identifier target", assignment.Left)
		}
	})

	t.Run("object destructuring target becomes an ObjectPattern", func(t *testing.T) {
		program := mustParse(t, "var map; ({ map } = require('lodash'));")
		assignment := collect(program, "AssignmentExpression")[0].(*AssignmentExpression)
		pattern, ok := assignment.Left.(*ObjectPattern)
		if !ok {
			t.Fatalf("Left is %T, want *ObjectPattern", assignment.Left)
		}
		if len(pattern.Properties) != 1 {
			t.Fatalf("Properties has %d entries, want 1", len(pattern.Properties))
		}
		if key := pattern.Properties[0].(*Property).Key.(*Identifier); key.Name != "map" {
			t.Errorf("Key = %q, want map", key.Name)
		}
	})

	t.Run("non assignment binary expression becomes Unknown", func(t *testing.T) {
		program := mustParse(t, "var v = left + right;")
		var kinds []string
		Walk(program, func(n Node) {
			if unknown, ok := n.(*Unknown); ok {
				kinds = append(kinds, unknown.Kind)
			}
		})
		found := false
		for _, kind := range kinds {
			if kind == "BinaryExpression" {
				found = true
			}
		}
		if !found {
			t.Errorf("Unknown kinds = %v, want one of them to be BinaryExpression", kinds)
		}
	})
}
