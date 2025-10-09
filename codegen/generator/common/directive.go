package common

const (
	DirectivePrefix          = "codegen"
	DirectiveCommentIgnore   = "ignore"
	DirectiveCommentGenerated = "generated"
)

// Direction represents the parsed directive information
type Direction struct {
	Ignore    bool `directive:"ignore"`
	Generated bool `directive:"generated"`
}

// MustIgnore returns true if the directive indicates the code should be ignored
func (d Direction) MustIgnore() bool {
	return d.Ignore
}

// IsGenerated returns true if the directive indicates the code is generated
func (d Direction) IsGenerated() bool {
	return d.Generated
}