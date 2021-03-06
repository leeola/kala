package q

import (
	"strings"

	"github.com/leeola/fixity/q/operator"
	"github.com/leeola/fixity/value"
	"github.com/mgutz/str"
)

// FromString produces a Query from the given string.
//
// Intended for constructing Queries from user input.
//
// TODO(leeola): support AND/OR by looking check if one of the parts equals
// AND/OR directly. Can also support -AND and -OR. Though i may have to
// implement my own parsing, to group ( and ), eg AND( ... ).
func FromString(s string) Query {
	parts := str.ToArgv(s)

	// the fieldless constraint is any parts that do not produce
	// another type of constraint. Ie, the resulting string.
	var fieldless []string

	var cs []Constraint
	for _, p := range parts {
		op, field, valueStr := splitPart(p)

		if op == "" && field == "" {
			fieldless = append(fieldless, valueStr)
			continue
		}

		switch op {
		case "eq":
			op = operator.Equal

		case "":
			// default empty ops to equal.
			//
			// in the future it should probably translate to some type of loose
			// operator, maybe fts?
			op = operator.Equal
		}

		v := value.String(valueStr)

		cs = append(cs, Constraint{
			Operator: op,
			Field:    &field,
			Value:    &v,
		})
	}

	if len(fieldless) != 0 {
		v := value.String(strings.Join(fieldless, " "))
		cs = append(cs, Constraint{
			Operator: operator.Equal,
			Value:    &v,
		})
	}

	if len(cs) == 1 {
		return New().Const(cs[0])
	}

	return New().And(cs...)
}

func splitPart(s string) (op, field, value string) {
	constStrs := strings.SplitN(s, ":", 3)
	switch len(constStrs) {
	case 1:
		// "value"
		value = constStrs[0]
	case 2:
		// "field:value"
		field = constStrs[0]
		value = constStrs[1]
	default:
		// "op:field:value"
		op = constStrs[0]
		field = constStrs[1]
		value = constStrs[2]
	}
	return
}
