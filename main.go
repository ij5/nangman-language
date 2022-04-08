package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var nmLexer = lexer.MustSimple([]lexer.Rule{
	{Name: "String", Pattern: `"(\\"|[^"])*"`, Action: nil},
	{Name: "Float", Pattern: `(\d*\.)\d+`, Action: nil},
	{Name: "Int", Pattern: `\d+`, Action: nil},
	{Name: "Ident", Pattern: `[ㄱ-ㅎ가-힣ㅏ-ㅣ]+`, Action: nil},
	{Name: "EOL", Pattern: `[\n]+`, Action: nil},
	{Name: "Whitespace", Pattern: `[ \t\r]+`, Action: nil},
	{Name: "Dot", Pattern: `\.`, Action: nil},
	{Name: "Operator", Pattern: `[\+\-\*\/]+`, Action: nil},
})

var parser = participle.MustBuild(
	&Statement{},
	participle.Lexer(nmLexer),
	participle.Elide("Whitespace"),
	// participle.UseLookahead(2),
)

type Operator int

const (
	OpMul Operator = iota
	OpDiv
	OpAdd
	OpSub
)

var operatorMap = map[string]Operator{"+": OpAdd, "-": OpSub, "*": OpMul, "/": OpDiv}

func (o *Operator) Capture(s []string) error {
	*o = operatorMap[s[0]]
	return nil
}

type Value struct {
	Float         *string     `  @Float`
	Int           *string     `| @Int`
	String        *string     `| @String`
	Subexpression *Expression `| "(" @@ ")"`
}

type Factor struct {
	Base *Value `@@`
}

type OpFactor struct {
	Operator *Operator `@("곱하기" | "나누기")`
	Factor   *Factor   `@@`
}

type Term struct {
	Left  *Factor     `@@`
	Right []*OpFactor `@@*`
}

type OpTerm struct {
	Operator *Operator `@("더하기" | "빼기")`
	Term     *Term     `@@`
}

type Expression struct {
	Left  *Term     `@@`
	Right []*OpTerm `@@*`
}

type Statement struct {
	Expr      *Expression `  @@`
	PrintFunc *Print      `| @@`
	ExitFunc  *Exit       `| @@`
	EOL       string      `| @EOF`
}

type Print struct {
	Name      string      `"나는" "그녀에게" "말했다" "."`
	Parameter *Expression `@@`
}

type Exit struct {
	Stmt string `"그녀는" "나를" ("차버렸다" | "찼다")`
}

func (s *Statement) Eval() {
	switch {
	case s.Expr != nil:
		s.Expr.Eval()
	case s.PrintFunc != nil:
		s.PrintFunc.Eval()
	case s.ExitFunc != nil:
		s.ExitFunc.Eval()
	default:
		return
	}
}

func (e *Expression) Eval() interface{} {
	l := e.Left.Eval()
	for _, r := range e.Right {
		l = r.Operator.Eval(l, r.Term.Eval())
	}
	return l
}

func (t *Term) Eval() interface{} {
	n := t.Left.Eval()
	for _, r := range t.Right {
		n = r.Operator.Eval(n, r.Factor.Eval())
	}
	return n
}

func (f *Factor) Eval() interface{} {
	b := f.Base.Eval()
	return b
}

func (v *Value) Eval() interface{} {
	switch {
	case v.Float != nil:
		f, err := strconv.ParseFloat(*v.Float, 64)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return f
	case v.Int != nil:
		f, err := strconv.ParseFloat(*v.Int, 64)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return f
	case v.String != nil:
		return (*v.String)[1 : len(*v.String)-1]
	default:
		return v.Subexpression.Eval()
	}
}

func (o Operator) Eval(l, r interface{}) interface{} {
	switch l.(type) {
	case string:
		fmt.Println("문자열은 연산할 수 없습니다.")
		return nil
	}
	switch r.(type) {
	case string:
		fmt.Println("문자열은 연산할 수 없습니다.")
		return nil
	}
	switch o {
	case OpMul:
		return l.(float64) * r.(float64)
	case OpDiv:
		return l.(float64) / r.(float64)
	case OpAdd:
		return l.(float64) + r.(float64)
	case OpSub:
		return l.(float64) - r.(float64)
	}
	panic("unsupported operator")
}

func (p *Print) Eval() {
	result := p.Parameter.Eval()
	switch result.(type) {
	case float64:
		fmt.Print("그녀가 말했다. \"")
		fmt.Print(result)
		fmt.Println("\"")
	case string:
		fmt.Print("그녀가 따라했다. \"")
		fmt.Print(result)
		fmt.Println("\"")
	}
}

func (e *Exit) Eval() {
	panic("프로그램이 종료되었습니다.")
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		input = strings.TrimSuffix(input, "\n")

		program := &Statement{}
		err = parser.ParseString("", input, program)
		if err != nil {
			fmt.Println(err)
			// json.NewEncoder(os.Stdout).Encode(program)
			continue
		}
		program.Eval()
	}

}
