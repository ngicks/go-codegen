package directive

const (
	DirectivePrefix = "codegen:"
)

const (
	DirectiveCommentIgnore    = "ignore"
	DirectiveCommentGenerated = "generated"
)

type Direction struct {
	Ignore    bool `directive:"ignore"`
	Generated bool `directive:"generated"`
}

func (d Direction) MustIgnore() bool {
	return d.Ignore || d.Generated
}

func (d Direction) IsGenerated() bool {
	return d.Generated
}