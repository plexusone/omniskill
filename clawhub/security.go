package clawhub

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SecurityScanner scans skill packages for security issues.
type SecurityScanner struct {
	// StrictMode fails on any warning when true.
	StrictMode bool

	// AllowedCommands is a list of allowed shell commands.
	AllowedCommands []string

	// BlockedPatterns are regex patterns that are not allowed.
	BlockedPatterns []*regexp.Regexp
}

// SecurityIssue represents a security finding.
type SecurityIssue struct {
	// Severity is the issue severity.
	Severity Severity

	// Type is the issue type.
	Type IssueType

	// File is the file where the issue was found.
	File string

	// Line is the line number (0 if not applicable).
	Line int

	// Description describes the issue.
	Description string

	// Recommendation suggests how to fix the issue.
	Recommendation string
}

// Severity represents the severity of a security issue.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// IssueType represents the type of security issue.
type IssueType string

const (
	IssueTypeSecretLeak    IssueType = "secret_leak"
	IssueTypeDangerousCmd  IssueType = "dangerous_command"
	IssueTypeNetworkAccess IssueType = "network_access"
	IssueTypeFileAccess    IssueType = "file_access"
	IssueTypeMalicious     IssueType = "malicious"
	IssueTypeUnsigned      IssueType = "unsigned"
)

// ScanResult contains the results of a security scan.
type ScanResult struct {
	// Issues is the list of security issues found.
	Issues []SecurityIssue

	// FileHashes contains SHA256 hashes of all files.
	FileHashes map[string]string

	// Passed indicates if the scan passed.
	Passed bool
}

// NewSecurityScanner creates a new security scanner with defaults.
func NewSecurityScanner() *SecurityScanner {
	return &SecurityScanner{
		AllowedCommands: []string{
			"echo", "cat", "grep", "sed", "awk", "jq",
			"curl", "wget", "git", "npm", "pip", "go",
		},
		BlockedPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(password|secret|api[_-]?key)\s*[:=]\s*['"][^'"]+['"]`),
			regexp.MustCompile(`(?i)sk-[a-zA-Z0-9]{48,}`),  // OpenAI keys
			regexp.MustCompile(`(?i)ghp_[a-zA-Z0-9]{36,}`), // GitHub tokens
			regexp.MustCompile(`rm\s+-rf\s+(/|~|\$HOME)`),
			regexp.MustCompile(`eval\s*\(`),
			regexp.MustCompile(`exec\s*\(`),
			regexp.MustCompile(`:\s*\(\s*\)\s*\{\s*:\|:\s*&\s*\}`), // Fork bomb
		},
	}
}

// Scan scans a skill directory for security issues.
func (s *SecurityScanner) Scan(dir string) (*ScanResult, error) {
	result := &ScanResult{
		FileHashes: make(map[string]string),
		Passed:     true,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)

		// Calculate file hash
		//nolint:gosec // G122: Path comes from filepath.Walk on a controlled directory
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		hash := sha256.Sum256(data)
		result.FileHashes[relPath] = hex.EncodeToString(hash[:])

		// Scan file content
		issues := s.scanFile(relPath, string(data))
		result.Issues = append(result.Issues, issues...)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scan directory: %w", err)
	}

	// Check if any critical or high issues were found
	for _, issue := range result.Issues {
		if issue.Severity == SeverityCritical || issue.Severity == SeverityHigh {
			result.Passed = false
			break
		}
		if s.StrictMode && issue.Severity == SeverityMedium {
			result.Passed = false
			break
		}
	}

	return result, nil
}

// scanFile scans a single file for security issues.
func (s *SecurityScanner) scanFile(path string, content string) []SecurityIssue {
	var issues []SecurityIssue

	lines := strings.Split(content, "\n")

	for i, line := range lines {
		lineNum := i + 1

		// Check for blocked patterns
		for _, pattern := range s.BlockedPatterns {
			if pattern.MatchString(line) {
				issues = append(issues, SecurityIssue{
					Severity:       s.severityForPattern(pattern),
					Type:           s.typeForPattern(pattern),
					File:           path,
					Line:           lineNum,
					Description:    fmt.Sprintf("Blocked pattern found: %s", pattern.String()),
					Recommendation: "Remove or obfuscate sensitive content",
				})
			}
		}

		// Check for dangerous shell commands
		if strings.Contains(path, ".md") || strings.Contains(path, ".sh") {
			for _, issue := range s.checkDangerousCommands(line, path, lineNum) {
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

// checkDangerousCommands checks for dangerous shell commands.
func (s *SecurityScanner) checkDangerousCommands(line, path string, lineNum int) []SecurityIssue {
	var issues []SecurityIssue

	dangerousPatterns := map[*regexp.Regexp]string{
		regexp.MustCompile(`chmod\s+777`):      "chmod 777 grants excessive permissions",
		regexp.MustCompile(`curl\s+.*\|\s*sh`): "Piping curl to shell is dangerous",
		regexp.MustCompile(`wget\s+.*\|\s*sh`): "Piping wget to shell is dangerous",
		regexp.MustCompile(`sudo\s+`):          "sudo usage requires elevated privileges",
		regexp.MustCompile(`>\s*/etc/`):        "Writing to system files is dangerous",
		regexp.MustCompile(`\$\([^)]+\)`):      "Command substitution found",
		regexp.MustCompile("`.+`"):             "Backtick command execution found",
	}

	for pattern, desc := range dangerousPatterns {
		if pattern.MatchString(line) {
			issues = append(issues, SecurityIssue{
				Severity:       SeverityMedium,
				Type:           IssueTypeDangerousCmd,
				File:           path,
				Line:           lineNum,
				Description:    desc,
				Recommendation: "Review and ensure the command is necessary and safe",
			})
		}
	}

	return issues
}

// severityForPattern returns the severity for a blocked pattern.
func (s *SecurityScanner) severityForPattern(pattern *regexp.Regexp) Severity {
	patternStr := pattern.String()
	if strings.Contains(patternStr, "password") || strings.Contains(patternStr, "secret") ||
		strings.Contains(patternStr, "sk-") || strings.Contains(patternStr, "ghp_") {
		return SeverityCritical
	}
	if strings.Contains(patternStr, "rm -rf") || strings.Contains(patternStr, "fork bomb") {
		return SeverityCritical
	}
	if strings.Contains(patternStr, "eval") || strings.Contains(patternStr, "exec") {
		return SeverityHigh
	}
	return SeverityMedium
}

// typeForPattern returns the issue type for a blocked pattern.
func (s *SecurityScanner) typeForPattern(pattern *regexp.Regexp) IssueType {
	patternStr := pattern.String()
	if strings.Contains(patternStr, "password") || strings.Contains(patternStr, "secret") ||
		strings.Contains(patternStr, "sk-") || strings.Contains(patternStr, "ghp_") {
		return IssueTypeSecretLeak
	}
	if strings.Contains(patternStr, "rm -rf") || strings.Contains(patternStr, "eval") ||
		strings.Contains(patternStr, "exec") || strings.Contains(patternStr, "fork bomb") {
		return IssueTypeMalicious
	}
	return IssueTypeDangerousCmd
}

// ValidateSignature validates a skill's cryptographic signature.
func ValidateSignature(manifest *Manifest, publicKey string) error {
	if manifest.Signature == "" {
		return fmt.Errorf("manifest is not signed")
	}

	// TODO: Implement actual signature validation using Ed25519 or similar
	// For now, this is a placeholder that always passes

	return nil
}
