// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package registry

import "strings"

// Capability represents a skill capability that agents can query for.
// Capabilities describe what a skill can do at a semantic level,
// allowing agents to find skills by need rather than by name.
type Capability string

// Standard capability constants.
// These represent common categories of functionality that skills provide.
const (
	// File system capabilities
	CapabilityFileRead   Capability = "file:read"
	CapabilityFileWrite  Capability = "file:write"
	CapabilityFileDelete Capability = "file:delete"
	CapabilityFileList   Capability = "file:list"

	// Network capabilities
	CapabilityHTTPRequest Capability = "http:request"
	CapabilityHTTPServer  Capability = "http:server"
	CapabilityWebSocket   Capability = "websocket"

	// Code execution capabilities
	CapabilityCodeExecute   Capability = "code:execute"
	CapabilityCodeAnalyze   Capability = "code:analyze"
	CapabilityCodeGenerate  Capability = "code:generate"
	CapabilityCodeFormat    Capability = "code:format"
	CapabilityCodeLint      Capability = "code:lint"
	CapabilityCodeTest      Capability = "code:test"
	CapabilityCodeRefactor  Capability = "code:refactor"
	CapabilityCodeTransform Capability = "code:transform"

	// Git/VCS capabilities
	CapabilityGitRead   Capability = "git:read"
	CapabilityGitWrite  Capability = "git:write"
	CapabilityGitCommit Capability = "git:commit"
	CapabilityGitPush   Capability = "git:push"

	// Database capabilities
	CapabilityDatabaseRead  Capability = "database:read"
	CapabilityDatabaseWrite Capability = "database:write"
	CapabilityDatabaseAdmin Capability = "database:admin"

	// Communication capabilities
	CapabilityEmailSend    Capability = "email:send"
	CapabilityEmailRead    Capability = "email:read"
	CapabilitySlackSend    Capability = "slack:send"
	CapabilitySlackRead    Capability = "slack:read"
	CapabilityCalendarRead Capability = "calendar:read"
	CapabilityCalendarEdit Capability = "calendar:edit"

	// Search capabilities
	CapabilitySearchWeb   Capability = "search:web"
	CapabilitySearchCode  Capability = "search:code"
	CapabilitySearchLocal Capability = "search:local"

	// Document capabilities
	CapabilityDocumentRead    Capability = "document:read"
	CapabilityDocumentWrite   Capability = "document:write"
	CapabilityDocumentConvert Capability = "document:convert"

	// AI/ML capabilities
	CapabilityAIChat       Capability = "ai:chat"
	CapabilityAIEmbed      Capability = "ai:embed"
	CapabilityAIComplete   Capability = "ai:complete"
	CapabilityAIImageGen   Capability = "ai:image_gen"
	CapabilityAITranscribe Capability = "ai:transcribe"

	// System capabilities
	CapabilitySystemShell   Capability = "system:shell"
	CapabilitySystemProcess Capability = "system:process"
	CapabilitySystemEnv     Capability = "system:env"
)

// Category returns the category part of the capability (before the colon).
func (c Capability) Category() string {
	s := string(c)
	if idx := strings.Index(s, ":"); idx > 0 {
		return s[:idx]
	}
	return s
}

// Action returns the action part of the capability (after the colon).
func (c Capability) Action() string {
	s := string(c)
	if idx := strings.Index(s, ":"); idx > 0 && idx < len(s)-1 {
		return s[idx+1:]
	}
	return ""
}

// String returns the string representation of the capability.
func (c Capability) String() string {
	return string(c)
}

// ParseCapability parses a string into a Capability.
// Returns the capability and true if valid, or empty and false if invalid.
func ParseCapability(s string) (Capability, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", false
	}
	// Accept both "category:action" and "category_action" formats
	s = strings.ReplaceAll(s, "_", ":")
	return Capability(s), true
}

// ParseCapabilities parses a slice of strings into capabilities.
// Invalid strings are skipped.
func ParseCapabilities(keywords []string) []Capability {
	var caps []Capability
	for _, k := range keywords {
		if cap, ok := ParseCapability(k); ok {
			caps = append(caps, cap)
		}
	}
	return caps
}

// CapabilityFromKeywords infers capabilities from skill keywords/tags.
// This maps common keywords to their corresponding capabilities.
func CapabilityFromKeywords(keywords []string) []Capability {
	var caps []Capability
	seen := make(map[Capability]bool)

	for _, kw := range keywords {
		kw = strings.ToLower(strings.TrimSpace(kw))
		var inferred []Capability

		switch kw {
		// File keywords
		case "file", "filesystem", "fs":
			inferred = []Capability{CapabilityFileRead, CapabilityFileWrite, CapabilityFileList}
		case "read", "reader":
			inferred = []Capability{CapabilityFileRead}
		case "write", "writer":
			inferred = []Capability{CapabilityFileWrite}

		// Network keywords
		case "http", "rest", "api":
			inferred = []Capability{CapabilityHTTPRequest}
		case "server", "serve":
			inferred = []Capability{CapabilityHTTPServer}
		case "websocket", "ws":
			inferred = []Capability{CapabilityWebSocket}

		// Code keywords
		case "code", "programming":
			inferred = []Capability{CapabilityCodeAnalyze, CapabilityCodeGenerate}
		case "execute", "exec", "run":
			inferred = []Capability{CapabilityCodeExecute}
		case "lint", "linter":
			inferred = []Capability{CapabilityCodeLint}
		case "format", "formatter":
			inferred = []Capability{CapabilityCodeFormat}
		case "test", "testing":
			inferred = []Capability{CapabilityCodeTest}
		case "refactor", "refactoring":
			inferred = []Capability{CapabilityCodeRefactor}

		// Git keywords
		case "git", "vcs", "version-control":
			inferred = []Capability{CapabilityGitRead, CapabilityGitWrite}
		case "commit":
			inferred = []Capability{CapabilityGitCommit}

		// Database keywords
		case "database", "db", "sql":
			inferred = []Capability{CapabilityDatabaseRead, CapabilityDatabaseWrite}
		case "query":
			inferred = []Capability{CapabilityDatabaseRead}

		// Communication keywords
		case "email", "mail":
			inferred = []Capability{CapabilityEmailSend, CapabilityEmailRead}
		case "slack":
			inferred = []Capability{CapabilitySlackSend, CapabilitySlackRead}
		case "calendar", "meeting":
			inferred = []Capability{CapabilityCalendarRead, CapabilityCalendarEdit}

		// Search keywords
		case "search":
			inferred = []Capability{CapabilitySearchWeb, CapabilitySearchCode}
		case "web-search", "websearch":
			inferred = []Capability{CapabilitySearchWeb}
		case "code-search", "grep":
			inferred = []Capability{CapabilitySearchCode}

		// Document keywords
		case "document", "doc", "docs":
			inferred = []Capability{CapabilityDocumentRead, CapabilityDocumentWrite}
		case "convert", "converter":
			inferred = []Capability{CapabilityDocumentConvert}

		// AI keywords
		case "ai", "llm", "ml":
			inferred = []Capability{CapabilityAIChat, CapabilityAIComplete}
		case "chat", "chatbot":
			inferred = []Capability{CapabilityAIChat}
		case "embedding", "embed":
			inferred = []Capability{CapabilityAIEmbed}
		case "image-generation", "dall-e", "stable-diffusion":
			inferred = []Capability{CapabilityAIImageGen}
		case "transcribe", "speech-to-text", "whisper":
			inferred = []Capability{CapabilityAITranscribe}

		// System keywords
		case "shell", "bash", "terminal":
			inferred = []Capability{CapabilitySystemShell}
		case "process", "processes":
			inferred = []Capability{CapabilitySystemProcess}
		}

		for _, cap := range inferred {
			if !seen[cap] {
				seen[cap] = true
				caps = append(caps, cap)
			}
		}
	}

	return caps
}

// MatchesCategory returns true if the capability belongs to the given category.
func (c Capability) MatchesCategory(category string) bool {
	return c.Category() == strings.ToLower(category)
}

// AllCapabilities returns all standard capabilities.
func AllCapabilities() []Capability {
	return []Capability{
		CapabilityFileRead, CapabilityFileWrite, CapabilityFileDelete, CapabilityFileList,
		CapabilityHTTPRequest, CapabilityHTTPServer, CapabilityWebSocket,
		CapabilityCodeExecute, CapabilityCodeAnalyze, CapabilityCodeGenerate,
		CapabilityCodeFormat, CapabilityCodeLint, CapabilityCodeTest,
		CapabilityCodeRefactor, CapabilityCodeTransform,
		CapabilityGitRead, CapabilityGitWrite, CapabilityGitCommit, CapabilityGitPush,
		CapabilityDatabaseRead, CapabilityDatabaseWrite, CapabilityDatabaseAdmin,
		CapabilityEmailSend, CapabilityEmailRead, CapabilitySlackSend, CapabilitySlackRead,
		CapabilityCalendarRead, CapabilityCalendarEdit,
		CapabilitySearchWeb, CapabilitySearchCode, CapabilitySearchLocal,
		CapabilityDocumentRead, CapabilityDocumentWrite, CapabilityDocumentConvert,
		CapabilityAIChat, CapabilityAIEmbed, CapabilityAIComplete,
		CapabilityAIImageGen, CapabilityAITranscribe,
		CapabilitySystemShell, CapabilitySystemProcess, CapabilitySystemEnv,
	}
}
