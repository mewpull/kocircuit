package eval

import (
	"fmt"

	. "github.com/kocircuit/kocircuit/lang/circuit/model"
	. "github.com/kocircuit/kocircuit/lang/go/kit/tree"
)

type Combiner interface {
	Interpret(Evaluator, *Func) Macro
	Combine(*Span, *Func, Arg, Return, StepResidues) (Effect, error)
}

type StepResidue struct {
	Span   *Span  `ko:"name=span"`
	Shape  Shape  `ko:"name=shape"`
	Effect Effect `ko:"name=effect"`
}

type StepResidues []*StepResidue

func (sr StepResidues) String() string {
	return Sprint(sr)
}

type IdentityCombiner struct{}

func (IdentityCombiner) Combine(_ *Span, _ *Func, _ Arg, _ Return, effect StepResidues) (Effect, error) {
	return effect, nil
}

func (IdentityCombiner) Interpret(eval Evaluator, f *Func) Macro {
	return &evalFixedFuncMacro{Func: f, Parent: eval}
}

// evalFixedFuncMacro is a macro that plays an underlying circuit function with the parent evaluator.
type evalFixedFuncMacro struct {
	Func   *Func
	Parent Evaluator
}

func (m *evalFixedFuncMacro) Splay() Tree {
	return Quote{String_: m.Help()}
}

func (m *evalFixedFuncMacro) MacroID() string { return m.Help() }

func (m *evalFixedFuncMacro) MacroSheathString() *string { return nil }

func (m *evalFixedFuncMacro) Label() string { return "evalfixed" }

func (m *evalFixedFuncMacro) Help() string {
	return fmt.Sprintf("Eval(%s)", m.Func.FullPath())
}

func (m *evalFixedFuncMacro) Doc() string {
	return m.Func.DocLong()
}

func (m *evalFixedFuncMacro) Invoke(span *Span, arg Arg) (Return, Effect, error) {
	if arg == nil {
		return m.Parent.EvalSeq(span, m.Func, nil)
	} else {
		return m.Parent.EvalSeq(span, m.Func, arg)
	}
}
