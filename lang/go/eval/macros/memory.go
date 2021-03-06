package macros

import (
	"fmt"
	"sync"

	. "github.com/kocircuit/kocircuit/lang/circuit/eval"
	. "github.com/kocircuit/kocircuit/lang/circuit/model"
	. "github.com/kocircuit/kocircuit/lang/go/eval"
	. "github.com/kocircuit/kocircuit/lang/go/eval/symbol"
	. "github.com/kocircuit/kocircuit/lang/go/kit/util"
)

type keyValues []*keyValue

func (kvs keyValues) Find(span *Span, key Symbol) *keyValue {
	for _, kv := range kvs {
		if key.Equal(span, kv.key) {
			return kv
		}
	}
	return nil
}

func (kvs keyValues) Remove(span *Span, key Symbol) keyValues {
	filtered := make(keyValues, 0, len(kvs))
	for _, kv := range kvs {
		if !key.Equal(span, kv.key) {
			filtered = append(filtered, kv)
		}
	}
	return filtered
}

func (kvs keyValues) Add(span *Span, key, value Symbol) keyValues {
	updated := make(keyValues, 0, len(kvs)+1)
	for _, kv := range kvs {
		if !key.Equal(span, kv.key) {
			updated = append(updated, kv)
		}
	}
	updated = append(updated, &keyValue{key: key, value: value})
	return updated
}

type memory struct {
	origin *Span
	sync.Mutex
	seen map[ID]keyValues
}

type keyValue struct {
	key   Symbol
	value Symbol
}

func newMemory(origin *Span) *memory {
	return &memory{origin: origin, seen: map[ID]keyValues{}}
}

func (m *memory) ID() ID {
	return m.origin.SpanID()
}

func (m *memory) Remember(span *Span, key, value Symbol) Symbol {
	m.Lock()
	defer m.Unlock()
	keyHash := key.Hash(span)
	oldKeyValues := m.seen[keyHash]
	if IsEmptySymbol(value) {
		if u := oldKeyValues.Remove(span, key); len(u) == 0 {
			delete(m.seen, keyHash)
		} else {
			m.seen[keyHash] = u
		}
	} else {
		m.seen[keyHash] = oldKeyValues.Add(span, key, value)
	}
	if old := oldKeyValues.Find(span, key); old == nil {
		return EmptySymbol{}
	} else {
		return old.value
	}
}

func (m *memory) Recall(span *Span, key Symbol) Symbol {
	m.Lock()
	defer m.Unlock()
	if keyValues, found := m.seen[key.Hash(span)]; found {
		if kv := keyValues.Find(span, key); kv != nil {
			return kv.value
		}
	}
	return EmptySymbol{}
}

func (m *memory) Flush(span *Span) (Symbol, error) {
	m.Lock()
	defer m.Unlock()
	kvElems := make(Symbols, 0, len(m.seen))
	for _, kvs := range m.seen {
		for _, kv := range kvs {
			kvElems = append(kvElems,
				MakeStructSymbol(
					FieldSymbols{
						&FieldSymbol{Name: "key", Value: kv.key},
						&FieldSymbol{Name: "value", Value: kv.value},
					},
				),
			)
		}
	}
	return MakeSeriesSymbol(span, kvElems)
}

func init() {
	RegisterEvalMacro("Memory", new(EvalMemoryMacro))
}

type EvalMemoryMacro struct{}

func (m EvalMemoryMacro) MacroID() string { return m.Help() }

func (m EvalMemoryMacro) Label() string { return "memory" }

func (m EvalMemoryMacro) MacroSheathString() *string { return PtrString("Memory") }

func (m EvalMemoryMacro) Help() string { return "Memory" }

func (m EvalMemoryMacro) Doc() string {
	return `Memory returns a new key-value memory.`
}

func (EvalMemoryMacro) Invoke(span *Span, arg Arg) (returns Return, effect Effect, err error) {
	m := newMemory(span)
	return MakeStructSymbol(
		FieldSymbols{
			{
				Name:  "name",
				Value: MakeBasicSymbol(span, m.ID().String()),
			},
			{
				Name:  "Remember",
				Value: MakeVarietySymbol(&evalRememberMacro{m}, nil),
			},
			{
				Name:  "Recall",
				Value: MakeVarietySymbol(&evalRecallMacro{m}, nil),
			},
			{
				Name:  "Flush",
				Value: MakeVarietySymbol(&evalFlushMacro{m}, nil),
			},
		},
	), nil, nil
}

// Remember
type evalRememberMacro struct {
	memory *memory
}

func (m evalRememberMacro) MacroID() string { return m.Help() }

func (m evalRememberMacro) Label() string { return "remember" }

func (m evalRememberMacro) MacroSheathString() *string { return PtrString("Remember") }

func (m evalRememberMacro) Help() string {
	return fmt.Sprintf("%v_Remember", m.memory.ID())
}

func (m evalRememberMacro) Doc() string {
	return `Remember(key, value) places the key-value pair in memory and returns the previously stored value for key.`
}

func (m evalRememberMacro) Invoke(span *Span, arg Arg) (returns Return, effect Effect, err error) {
	a := arg.(*StructSymbol)
	return m.memory.Remember(span, a.Walk("key"), a.Walk("value")), nil, nil
}

// Recall
type evalRecallMacro struct {
	memory *memory
}

func (m evalRecallMacro) MacroID() string { return m.Help() }

func (m evalRecallMacro) Label() string { return "recall" }

func (m evalRecallMacro) MacroSheathString() *string { return PtrString("Recall") }

func (m evalRecallMacro) Help() string {
	return fmt.Sprintf("%v_Recall", m.memory.ID())
}

func (m evalRecallMacro) Doc() string {
	return `Recall(key, otherwise) returns the previously stored value for key or the value otherwise.`
}

func (m evalRecallMacro) Invoke(span *Span, arg Arg) (returns Return, effect Effect, err error) {
	a := arg.(*StructSymbol)
	if recalled := m.memory.Recall(span, a.Walk("key")); IsEmptySymbol(recalled) {
		return a.Walk("otherwise"), nil, nil
	} else {
		return recalled, nil, nil
	}
}

// Flush
type evalFlushMacro struct {
	memory *memory
}

func (m evalFlushMacro) MacroID() string { return m.Help() }

func (m evalFlushMacro) Label() string { return "flush" }

func (m evalFlushMacro) MacroSheathString() *string { return PtrString("Flush") }

func (m evalFlushMacro) Help() string {
	return fmt.Sprintf("%v_Flush", m.memory.ID())
}

func (m evalFlushMacro) Doc() string {
	return `Flush returns the sequence of (key, value) pairs stored in memory.`
}

func (m evalFlushMacro) Invoke(span *Span, arg Arg) (returns Return, effect Effect, err error) {
	if flushed, err := m.memory.Flush(span); err != nil {
		return nil, nil, span.Errorf(err, "flushing key-value memory")
	} else {
		return flushed, nil, nil
	}
}
