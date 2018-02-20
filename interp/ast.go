package interp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

type Kind int

const (
	Undef = iota
	ArrayType
	AssignStmt
	BasicLit
	BinaryExpr
	BlockStmt
	BranchStmt
	Break
	CallExpr
	CompositeLit
	Continue
	ExprStmt
	Fallthrough
	Field
	FieldList
	File
	For0         // for {}
	For1         // for cond {}
	For2         // for init; cond; {}
	For3         // for ; cond; post {}
	For4         // for init; cond; post {}
	ForRangeStmt // for range
	ForStmt
	FuncDecl
	FuncType
	Goto
	Ident
	If0 // if cond {}
	If1 // if cond {} else {}
	If2 // if init; cond {}
	If3 // if init; cond {} else {}
	IfStmt
	IncDecStmt
	IndexExpr
	ParenExpr
	RangeStmt
	ReturnStmt
)

var kinds = [...]string{
	Undef:        "Undef",
	ArrayType:    "ArrayType",
	AssignStmt:   "AssignStmt",
	BasicLit:     "BasicLit",
	BinaryExpr:   "BinaryExpr",
	BlockStmt:    "BlockStmt",
	BranchStmt:   "BranchStmt",
	Break:        "Break",
	CallExpr:     "CallExpr",
	CompositeLit: "CompositLit",
	Continue:     "Continue",
	ExprStmt:     "ExprStmt",
	Field:        "Field",
	FieldList:    "FieldList",
	File:         "File",
	For0:         "For0",
	For1:         "For1",
	For2:         "For2",
	For3:         "For3",
	For4:         "For4",
	ForRangeStmt: "ForRangeStmt",
	ForStmt:      "ForStmt",
	FuncDecl:     "FuncDecl",
	FuncType:     "FuncType",
	Goto:         "Goto",
	Ident:        "Ident",
	If0:          "If0",
	If1:          "If1",
	If2:          "If2",
	If3:          "If3",
	IfStmt:       "IfStmt",
	IncDecStmt:   "IncDecStmt",
	IndexExpr:    "IndexExpr",
	ParenExpr:    "ParenExpr",
	RangeStmt:    "RangeStmt",
	ReturnStmt:   "ReturnStmt",
}

func (k Kind) String() string {
	s := ""
	if 0 <= k && k <= Kind(len(kinds)) {
		s = kinds[k]
	}
	if s == "" {
		s = "kind(" + strconv.Itoa(int(k)) + ")"
	}
	return s
}

type Action int

const (
	Nop = iota
	ArrayLit
	Assign
	AssignX
	Add
	And
	Call
	Dec
	Equal
	Greater
	GetIndex
	Inc
	Lower
	Println
	Range
	Return
	Sub
)

var actions = [...]string{
	Nop:      "nop",
	ArrayLit: "arraylit",
	Assign:   "=",
	AssignX:  "=",
	Add:      "+",
	And:      "&",
	Call:     "call",
	Dec:      "--",
	Equal:    "==",
	Greater:  ">",
	GetIndex: "getindex",
	Inc:      "++",
	Lower:    "<",
	Println:  "println",
	Range:    "range",
	Return:   "return",
	Sub:      "-",
}

func (a Action) String() string {
	s := ""
	if 0 <= a && a <= Action(len(actions)) {
		s = actions[a]
	}
	if s == "" {
		s = "action(" + strconv.Itoa(int(a)) + ")"
	}
	return s
}

type Def map[string]*Node // map of defined symbols

// Ast(src) parses src string containing Go code and generates the corresponding AST.
// The AST root node is returned.
func Ast(src string) (*Node, Def) {
	var def Def = make(map[string]*Node)
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "sample.go", src, 0)
	if err != nil {
		panic(err)
	}
	//ast.Print(fset, f)

	index := 0
	var root *Node
	var anc *Node
	var st nodestack
	// Populate our own private AST from Go parser AST.
	// A stack of ancestor nodes is used to keep track of curent ancestor for each depth level
	ast.Inspect(f, func(node ast.Node) bool {
		anc = st.top()
		switch a := node.(type) {
		case nil:
			anc = st.pop()

		case *ast.ArrayType:
			st.push(addChild(&root, anc, &index, ArrayType, Nop, &node))

		case *ast.AssignStmt:
			var action Action
			if len(a.Lhs) > 1 && len(a.Rhs) == 1 {
				action = AssignX
			} else {
				action = Assign
			}
			st.push(addChild(&root, anc, &index, AssignStmt, action, &node))

		case *ast.BasicLit:
			n := addChild(&root, anc, &index, BasicLit, Nop, &node)
			// FIXME: values must be converted to int or float if possible
			n.ident = a.Value
			if v, err := strconv.ParseInt(a.Value, 0, 0); err == nil {
				n.val = int(v)
			} else {
				n.val = a.Value
			}
			st.push(n)

		case *ast.BinaryExpr:
			var action Action
			switch a.Op {
			case token.ADD:
				action = Add
			case token.AND:
				action = And
			case token.EQL:
				action = Equal
			case token.GTR:
				action = Greater
			case token.LSS:
				action = Lower
			case token.SUB:
				action = Sub
			}
			st.push(addChild(&root, anc, &index, BinaryExpr, action, &node))

		case *ast.BlockStmt:
			st.push(addChild(&root, anc, &index, BlockStmt, Nop, &node))

		case *ast.BranchStmt:
			var kind Kind
			switch a.Tok {
			case token.BREAK:
				kind = Break
			case token.CONTINUE:
				kind = Continue
			}
			st.push(addChild(&root, anc, &index, kind, Nop, &node))

		case *ast.CallExpr:
			st.push(addChild(&root, anc, &index, CallExpr, Call, &node))

		case *ast.CompositeLit:
			st.push(addChild(&root, anc, &index, CompositeLit, ArrayLit, &node))

		case *ast.ExprStmt:
			st.push(addChild(&root, anc, &index, ExprStmt, Nop, &node))

		case *ast.Field:
			st.push(addChild(&root, anc, &index, Field, Nop, &node))

		case *ast.FieldList:
			st.push(addChild(&root, anc, &index, FieldList, Nop, &node))

		case *ast.File:
			st.push(addChild(&root, anc, &index, File, Nop, &node))

		case *ast.ForStmt:
			var kind Kind
			if a.Cond == nil {
				kind = For0
			} else {
				if a.Init == nil && a.Post == nil {
					kind = For1
				} else if a.Init != nil && a.Post == nil {
					kind = For2
				} else if a.Init == nil && a.Post != nil {
					kind = For3
				} else {
					kind = For4
				}
			}
			st.push(addChild(&root, anc, &index, kind, Nop, &node))

		case *ast.FuncDecl:
			n := addChild(&root, anc, &index, FuncDecl, Nop, &node)
			// Add func name to definitions
			def[a.Name.Name] = n
			st.push(n)

		case *ast.FuncType:
			st.push(addChild(&root, anc, &index, FuncType, Nop, &node))

		case *ast.Ident:
			n := addChild(&root, anc, &index, Ident, Nop, &node)
			n.ident = a.Name
			st.push(n)

		case *ast.IfStmt:
			var kind Kind
			if a.Init == nil && a.Else == nil {
				kind = If0
			} else if a.Init == nil && a.Else != nil {
				kind = If1
			} else if a.Else == nil {
				kind = If2
			} else {
				kind = If3
			}
			st.push(addChild(&root, anc, &index, kind, Nop, &node))

		case *ast.IncDecStmt:
			var action Action
			switch a.Tok {
			case token.INC:
				action = Inc
			case token.DEC:
				action = Dec
			}
			st.push(addChild(&root, anc, &index, IncDecStmt, action, &node))

		case *ast.IndexExpr:
			st.push(addChild(&root, anc, &index, IndexExpr, GetIndex, &node))

		case *ast.ParenExpr:
			st.push(addChild(&root, anc, &index, ParenExpr, Nop, &node))

		case *ast.RangeStmt:
			// Insert a missing ForRangeStmt for AST correctness
			n := addChild(&root, anc, &index, ForRangeStmt, Nop, nil)
			st.push(addChild(&root, n, &index, RangeStmt, Range, &node))

		case *ast.ReturnStmt:
			st.push(addChild(&root, anc, &index, ReturnStmt, Return, &node))

		default:
			fmt.Printf("Unknown kind for %T\n", a)
			st.push(addChild(&root, anc, &index, Undef, Nop, &node))
		}
		return true
	})
	return root, def
}

func addChild(root **Node, anc *Node, index *int, kind Kind, action Action, anode *ast.Node) *Node {
	*index++
	var i interface{}
	//n := &Node{anc: anc, index: *index, kind: kind, action: action, anode: anode, val: &i, run: builtin[action]}
	n := &Node{anc: anc, index: *index, kind: kind, action: action, val: &i, run: builtin[action]}
	n.Start = n
	if anc == nil {
		*root = n
	} else {
		anc.Child = append(anc.Child, n)
	}
	return n
}

type nodestack []*Node

func (s *nodestack) push(v *Node) {
	*s = append(*s, v)
}

func (s *nodestack) pop() *Node {
	l := len(*s) - 1
	res := (*s)[l]
	*s = (*s)[:l]
	return res
}

func (s *nodestack) top() *Node {
	l := len(*s)
	if l > 0 {
		return (*s)[l-1]
	}
	return nil
}
