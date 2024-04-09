package enumlike

//floating comment
//enum:target=Floating;variants=foo,bar,baz;group=Ahh:foo,baz

type Floating string

//enum:variants=foo,bar,baz;group=Ahh:foo,baz
type Str string

type DefinedStr Str

//enum:group=Aye:var1,const2;group=Nay:var2,const1,const2
type Predefined string

var (
	Var1 Predefined = "var1"
	Var2 Predefined = "var2"
)

const (
	Const1 Predefined = "const1"
	Const2 Predefined = "const2"
)
