package providers

import (
	"sort"
	"strings"
)

const (
	Claude      = "claude"
	Codex       = "codex"
	Antigravity = "antigravity"
	Hermes      = "hermes"
)

var canonicalProviderIDs = []string{
	Claude,
	Codex,
	Antigravity,
	Hermes,
}

var providerAliases = map[string]string{
	"agy": Antigravity,
}

var managedTargetRoots = map[string]string{
	Claude:      ".claude",
	Codex:       ".codex",
	Antigravity: ".config/agy",
	Hermes:      ".hermes",
}

func CanonicalID(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, providerID := range canonicalProviderIDs {
		if value == providerID {
			return providerID, true
		}
	}
	canonical, exists := providerAliases[value]
	return canonical, exists
}

func CanonicalIDs() []string {
	return append([]string{}, canonicalProviderIDs...)
}

func ManagedTargetRoot(providerID string) (string, bool) {
	canonical, exists := CanonicalID(providerID)
	if !exists {
		return "", false
	}
	root, exists := managedTargetRoots[canonical]
	return root, exists
}

func LegacyAliases(providerID string) []string {
	canonical, exists := CanonicalID(providerID)
	if !exists {
		return nil
	}
	var aliases []string
	for alias, target := range providerAliases {
		if target == canonical {
			aliases = append(aliases, alias)
		}
	}
	sort.Strings(aliases)
	return aliases
}
