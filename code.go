package meta

import (
	"fmt"
	"go/format"
	"go/types"
)

type Code struct {
	Imports []*types.Package
	Code    []byte
	err     error
}

func (c Code) Errorf(format string, a ...interface{}) Code {
	c.err = fmt.Errorf(format, a...)
	return c
}

// func (c *Code) Reset() {
// 	c.Code = c.Code[:0]
// 	c.Imports = c.Imports[:0]
// 	c.err = nil
// }

func Printf(src string, args ...interface{}) (code Code) {
	return code.Printf(src, args...)
}

// func (c Code) Conditional(block Code, conditions ...Code) Code {
// 	if len(conditions) == 0 || block.err != nil {
// 		return block
// 	}
// 	c.Imports = append(c.Imports, block.Imports...)
// 	all := ""
// 	for i := range conditions {
// 		cond := &conditions[i]
// 		if cond.err != nil {
// 			return *cond
// 		}
// 		if len(cond.Code) == 0 {
// 			continue
// 		}
// 		c.Imports = append(c.Imports, cond.Imports...)
// 		if all != "" {
// 			all += " && "
// 		}
// 		all += string(cond.Code)
// 	}
// 	if all == "" {
// 		return block
// 	}
// 	return c.Printf(`if %s {
// 		%s
// 	}`, all, block)

// }
func (c Code) String() string {
	return string(c.Code)
}

func (c Code) Err() error {
	return c.err
}
func (c Code) Error(err error) Code {
	c.err = err
	return c
}

func (c Code) Format() Code {
	c.Code, c.err = format.Source(c.Code)
	return c
}

func (c Code) Append(cc Code) Code {
	c.Code = append(c.Code, cc.Code...)
	c.Imports = append(c.Imports, cc.Imports...)
	if c.err == nil {
		c.err = cc.err
	}
	return c
}

func (c Code) Print(args ...interface{}) Code {
	c.Code = append(c.Code, fmt.Sprint(args...)...)
	return c
}

func (c Code) Println(args ...interface{}) Code {
	c.Code = append(c.Code, fmt.Sprintln(args...)...)
	return c
}

func (c Code) QPrintf(q types.Qualifier, format string, args ...interface{}) Code {
	c = c.Import(args...)
	if c.err != nil {
		return c
	}
	for i, a := range args {
		if t, ok := a.(types.Type); ok {
			args[i] = types.TypeString(t, q)
		}
	}
	c.Code = append(c.Code, fmt.Sprintf(format, args...)...)
	return c
}
func (c Code) Printf(format string, args ...interface{}) Code {
	c.Code = append(c.Code, fmt.Sprintf(format, args...)...)
	return c
}

func (c Code) Import(args ...interface{}) Code {
	for _, a := range args {
		switch a := a.(type) {
		case Code:
			if c.err = a.Err(); c.err != nil {
				return c
			}
			c.Imports = append(c.Imports, a.Imports...)
		case *Code:
			if a == nil {
				return c
			}
			if c.err = a.Err(); c.err != nil {
				return c
			}
			c.Imports = append(c.Imports, a.Imports...)
		default:
			c.Imports = TypeImports(c.Imports, a)
		}
	}
	return c
}

func (c *Code) Write(p []byte) (int, error) {
	c.Code = append(c.Code, p...)
	return len(p), nil
}
