package autoimpl

import (
	"fmt"

	"github.com/ngicks/go-codegen/codegen/generator/cloner"
	"github.com/ngicks/go-codegen/codegen/pkg/suffixwriter"
	"golang.org/x/tools/go/packages"
)

// DefaultRegistry returns the default generator registry with all built-in generators.
func DefaultRegistry() GeneratorRegistry {
	return GeneratorRegistry{
		"cloner":        invokeCloner,
		"und:patch":     invokeUndPatch,
		"und:plain":     invokeUndPlain,
		"und:validator": invokeUndValidator,
	}
}

// invokeCloner invokes the cloner generator for marked types.
func invokeCloner(marked []MarkedType, pkgs []*packages.Package, config map[string]string) error {
	if len(marked) == 0 {
		return nil
	}

	// Create cloner config from marker config
	cfg := &cloner.Config{
		MatcherConfig: &cloner.MatcherConfig{},
	}

	// Map configuration options
	if noCopy, ok := config["no-copy"]; ok {
		switch noCopy {
		case "ignore":
			cfg.MatcherConfig.NoCopyHandle = cloner.CopyHandleIgnore
		case "disallow":
			cfg.MatcherConfig.NoCopyHandle = cloner.CopyHandleDisallow
		case "copy":
			cfg.MatcherConfig.NoCopyHandle = cloner.CopyHandleCopyPointer
		}
	}

	if chanHandle, ok := config["chan"]; ok {
		switch chanHandle {
		case "ignore":
			cfg.MatcherConfig.ChannelHandle = cloner.CopyHandleIgnore
		case "disallow":
			cfg.MatcherConfig.ChannelHandle = cloner.CopyHandleDisallow
		case "copy":
			cfg.MatcherConfig.ChannelHandle = cloner.CopyHandleCopyPointer
		case "make":
			cfg.MatcherConfig.ChannelHandle = cloner.CopyHandleMake
		}
	}

	if funcHandle, ok := config["func"]; ok {
		switch funcHandle {
		case "ignore":
			cfg.MatcherConfig.FuncHandle = cloner.CopyHandleIgnore
		case "disallow":
			cfg.MatcherConfig.FuncHandle = cloner.CopyHandleDisallow
		case "copy":
			cfg.MatcherConfig.FuncHandle = cloner.CopyHandleCopyPointer
		}
	}

	if interfaceHandle, ok := config["interface"]; ok {
		switch interfaceHandle {
		case "ignore":
			cfg.MatcherConfig.InterfaceHandle = cloner.CopyHandleIgnore
		case "copy":
			cfg.MatcherConfig.InterfaceHandle = cloner.CopyHandleCopyPointer
		}
	}

	// TODO: We'll need to integrate with the existing cloner.Generate
	// For now, this is a placeholder that will be implemented properly
	// once we have the full context of marked types
	return fmt.Errorf("cloner invocation not yet implemented")
}

// invokeUndPatch invokes the und:patch generator for marked types.
func invokeUndPatch(marked []MarkedType, pkgs []*packages.Package, config map[string]string) error {
	if len(marked) == 0 {
		return nil
	}

	// TODO: Implement und patch invocation using undgen.GeneratePatcher
	return fmt.Errorf("und:patch invocation not yet implemented")
}

// invokeUndPlain invokes the und:plain generator for marked types.
func invokeUndPlain(marked []MarkedType, pkgs []*packages.Package, config map[string]string) error {
	if len(marked) == 0 {
		return nil
	}

	// TODO: Implement und plain invocation using undgen.GeneratePlain
	return fmt.Errorf("und:plain invocation not yet implemented")
}

// invokeUndValidator invokes the und:validator generator for marked types.
func invokeUndValidator(marked []MarkedType, pkgs []*packages.Package, config map[string]string) error {
	if len(marked) == 0 {
		return nil
	}

	// TODO: Implement und validator invocation using undgen.GenerateValidator
	return fmt.Errorf("und:validator invocation not yet implemented")
}

// Dispatch runs the appropriate generators for all marked types.
func Dispatch(cfg *DispatchConfig, marked []MarkedType, pkgs []*packages.Package, writer *suffixwriter.Writer) error {
	// Group marked types by generator
	byGenerator := make(map[string][]MarkedType)
	for _, m := range marked {
		for _, gen := range m.Generators {
			// Apply filters
			if len(cfg.OnlyGenerators) > 0 {
				found := false
				for _, only := range cfg.OnlyGenerators {
					if gen == only {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Check skip list
			skip := false
			for _, skipGen := range cfg.SkipGenerators {
				if gen == skipGen {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			byGenerator[gen] = append(byGenerator[gen], m)
		}
	}

	// Dispatch to each generator
	for genType, types := range byGenerator {
		genFunc, ok := cfg.GeneratorRegistry[genType]
		if !ok {
			return fmt.Errorf("unknown generator type: %s", genType)
		}

		// Extract config for this generator (use config from first marked type)
		var config map[string]string
		if len(types) > 0 {
			config = types[0].Config
		}

		if err := genFunc(types, pkgs, config); err != nil {
			return fmt.Errorf("generator %s failed: %w", genType, err)
		}
	}

	return nil
}
