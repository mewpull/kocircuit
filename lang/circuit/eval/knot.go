package eval

import (
	"fmt"

	. "github.com/kocircuit/kocircuit/lang/circuit/model"
	. "github.com/kocircuit/kocircuit/lang/circuit/syntax"
	. "github.com/kocircuit/kocircuit/lang/go/kit/tree"
)

type Arg interface {
	Shape
}

type Return interface {
	Shape
}

type Field struct {
	Name  string `ko:"name=name"` // step label or arg name
	Shape Shape  `ko:"name=shape"`
}

func (f Field) String() string { return Sprint(f) }

type Fields []Field

func (v Fields) String() string { return Sprint(v) }

func (v Fields) IsEmpty() bool { return len(v) == 0 }

// Link implements Shape.Select. For eval_test.go only.
func (v Fields) Link(span *Span, name string, monadic bool) (Shape, Effect, error) {
	if monadic {
		if s, eff, err := v.Select(span, []string{NoLabel}); err != nil {
			return nil, nil, err
		} else if s != (Empty{}) {
			return s, eff, nil
		}
	}
	return v.Select(span, []string{name})
}

// Select implements Shape.Select. For eval_test.go only.
func (v Fields) Select(span *Span, path Path) (Shape, Effect, error) {
	if len(path) == 0 {
		return v, nil, nil
	}
	projection := v.RestrictTo(path[0])
	switch len(projection) {
	case 0:
		return Empty{}, nil, nil
	case 1:
		return projection[0].Shape.Select(span, path[1:])
	}
	if len(path) > 1 {
		return nil, nil, span.Errorf(nil, "selecting into a sequence")
	}
	return projection, nil, nil
}

// Augment implements Shape.Augment.
func (v Fields) Augment(span *Span, _ Fields) (Shape, Effect, error) {
	return nil, nil, span.Errorf(nil, "cannot augment fields")
}

// Invoke implements Shape.Invoke.
func (v Fields) Invoke(span *Span) (Shape, Effect, error) {
	return nil, nil, span.Errorf(nil, "cannot invoke fields")
}

func (v Fields) Fields() []Field { return v }

func (v Fields) Names() []string {
	n := map[string]bool{}
	r := []string{}
	for _, f := range v {
		if !n[f.Name] {
			n[f.Name] = true
			r = append(r, f.Name)
		}
	}
	return r
}

func (v Fields) FieldGroup() [][]Field {
	r := [][]Field{}
	for _, n := range v.Names() {
		r = append(r, v.RestrictTo(n))
	}
	return r
}

func (v Fields) RestrictTo(name string) Fields {
	r := Fields{}
	for _, f := range v {
		if f.Name == name {
			r = append(r, f)
		}
	}
	return r
}

func (v Fields) StringField(label string) (string, error) {
	g := v.RestrictTo(label)
	if len(g) != 1 {
		return "", fmt.Errorf("not a singleton (got %d) field", len(g))
	}
	s, ok := g[0].Shape.(String)
	if !ok {
		return "", fmt.Errorf("not a string field (type is %T)", g[0].Shape)
	}
	return s.Value_, nil
}
