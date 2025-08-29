package filterast

// filterGenDecl
type Decl1 struct{}

type Decl2 struct{}

// filterGenDecl
type (
	Decl3 struct{}
)

type (
	// filterTypeSpec
	Decl4 struct{}
	Decl5 struct{}
)
