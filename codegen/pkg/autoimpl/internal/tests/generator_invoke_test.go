package tests

import (
	"testing"

	"github.com/ngicks/go-codegen/codegen/pkg/autoimpl"
)

func TestClonerInvocation(t *testing.T) {
	// Test that cloner generator is correctly invoked

	marked := []autoimpl.MarkedType{
		{
			TypeName:   "TestStruct",
			Generators: []string{"cloner"},
			Config: map[string]string{
				"no-copy": "copy",
			},
		},
	}

	// TODO: Test invokeCloner
	// Should configure cloner.Config correctly based on marker config
	// Should call cloner.Generate

	_ = marked
}

func TestUndPatchInvocation(t *testing.T) {
	// Test that und:patch generator is correctly invoked

	marked := []autoimpl.MarkedType{
		{
			TypeName:   "TestStruct",
			Generators: []string{"und:patch"},
			Config:     map[string]string{},
		},
	}

	// TODO: Test invokeUndPatch
	// Should configure undgen.Config with KindPatch
	// Should call undgen.GeneratePatcher

	_ = marked
}

func TestUndPlainInvocation(t *testing.T) {
	// Test that und:plain generator is correctly invoked

	marked := []autoimpl.MarkedType{
		{
			TypeName:   "TestStruct",
			Generators: []string{"und:plain"},
			Config:     map[string]string{},
		},
	}

	// TODO: Test invokeUndPlain
	_ = marked
}

func TestUndValidatorInvocation(t *testing.T) {
	// Test that und:validator generator is correctly invoked

	marked := []autoimpl.MarkedType{
		{
			TypeName:   "TestStruct",
			Generators: []string{"und:validator"},
			Config:     map[string]string{},
		},
	}

	// TODO: Test invokeUndValidator
	_ = marked
}

func TestClonerConfigMapping(t *testing.T) {
	// Test that marker config is correctly mapped to cloner.Config

	tests := []struct {
		name         string
		markerConfig map[string]string
		// We'll validate the resulting cloner.Config
	}{
		{
			name: "no-copy:copy",
			markerConfig: map[string]string{
				"no-copy": "copy",
			},
		},
		{
			name: "chan:make",
			markerConfig: map[string]string{
				"chan": "make",
			},
		},
		{
			name: "multiple options",
			markerConfig: map[string]string{
				"no-copy":   "copy",
				"chan":      "make",
				"func":      "copy",
				"interface": "copy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Test config mapping
			_ = tt
		})
	}
}
