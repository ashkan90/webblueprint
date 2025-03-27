package security

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// PermissionLevel represents the security level for blueprint execution
type PermissionLevel string

const (
	PermissionLevelRestricted PermissionLevel = "restricted" // Most restricted mode
	PermissionLevelStandard   PermissionLevel = "standard"   // Standard security
	PermissionLevelTrusted    PermissionLevel = "trusted"    // Trusted blueprints
	PermissionLevelAdmin      PermissionLevel = "admin"      // Administrative access
)

// SecurityChecker validates potentially dangerous operations
type SecurityChecker struct {
	permissionLevel PermissionLevel
	allowedDomains  map[string]bool
	blockedDomains  map[string]bool
	allowedPaths    map[string]bool
	blockedPaths    map[string]bool
	nodePermissions map[string]map[PermissionLevel]bool
	mutex           sync.RWMutex
}

// NewSecurityChecker creates a new security checker with the specified permission level
func NewSecurityChecker(permLevel PermissionLevel) *SecurityChecker {
	checker := &SecurityChecker{
		permissionLevel: permLevel,
		allowedDomains:  make(map[string]bool),
		blockedDomains:  make(map[string]bool),
		allowedPaths:    make(map[string]bool),
		blockedPaths:    make(map[string]bool),
		nodePermissions: make(map[string]map[PermissionLevel]bool),
	}

	// Initialize with default settings
	checker.setupDefaultPermissions()

	return checker
}

// setupDefaultPermissions configures default security settings
func (sc *SecurityChecker) setupDefaultPermissions() {
	// Node permissions by type and permission level
	sc.nodePermissions = map[string]map[PermissionLevel]bool{
		// Nodes that perform network operations
		"http-request": {
			PermissionLevelRestricted: false,
			PermissionLevelStandard:   true,
			PermissionLevelTrusted:    true,
			PermissionLevelAdmin:      true,
		},
		"websocket": {
			PermissionLevelRestricted: false,
			PermissionLevelStandard:   true,
			PermissionLevelTrusted:    true,
			PermissionLevelAdmin:      true,
		},
		// Nodes that interact with the file system
		"file-read": {
			PermissionLevelRestricted: false,
			PermissionLevelStandard:   false,
			PermissionLevelTrusted:    true,
			PermissionLevelAdmin:      true,
		},
		"file-write": {
			PermissionLevelRestricted: false,
			PermissionLevelStandard:   false,
			PermissionLevelTrusted:    true,
			PermissionLevelAdmin:      true,
		},
		// Nodes that execute system commands
		"execute-command": {
			PermissionLevelRestricted: false,
			PermissionLevelStandard:   false,
			PermissionLevelTrusted:    false,
			PermissionLevelAdmin:      true,
		},
		// All other nodes are allowed by default
		"*": {
			PermissionLevelRestricted: true,
			PermissionLevelStandard:   true,
			PermissionLevelTrusted:    true,
			PermissionLevelAdmin:      true,
		},
	}

	// Add some common blocked domains
	for _, domain := range []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"0.0.0.0",
		"169.254.",
		"192.168.",
		"10.",
		"172.16.",
		"172.17.",
		"172.18.",
		"172.19.",
		"172.20.",
		"172.21.",
		"172.22.",
		"172.23.",
		"172.24.",
		"172.25.",
		"172.26.",
		"172.27.",
		"172.28.",
		"172.29.",
		"172.30.",
		"172.31.",
	} {
		sc.blockedDomains[domain] = true
	}

	// Add some common blocked file paths
	for _, path := range []string{
		"/etc/",
		"/var/",
		"/root/",
		"/home/",
		"/proc/",
		"/sys/",
		"/dev/",
		".ssh/",
		".aws/",
		".env",
	} {
		sc.blockedPaths[path] = true
	}
}

// IsNodeAllowed checks if a node type is allowed to execute
func (sc *SecurityChecker) IsNodeAllowed(nodeType string, permLevel PermissionLevel) bool {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	if nodePermMap, exists := sc.nodePermissions[nodeType]; exists {
		if allowed, exists := nodePermMap[permLevel]; exists {
			return allowed
		}
	}

	// Fall back to the wildcard permissions
	if wildcardMap, exists := sc.nodePermissions["*"]; exists {
		if allowed, exists := wildcardMap[permLevel]; exists {
			return allowed
		}
	}

	// Default to false for unknown node types or permission levels
	return false
}

// IsDomainAllowed checks if a network domain is allowed
func (sc *SecurityChecker) IsDomainAllowed(domain string) bool {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	domain = strings.ToLower(domain)

	// Check if the domain is explicitly allowed
	if _, exists := sc.allowedDomains[domain]; exists {
		return true
	}

	// Check if the domain is explicitly blocked
	for blocked := range sc.blockedDomains {
		if strings.Contains(domain, blocked) {
			return false
		}
	}

	// Default to allowed for standard and higher permission levels
	return sc.permissionLevel != PermissionLevelRestricted
}

// IsFilePathAllowed checks if a file path is allowed
func (sc *SecurityChecker) IsFilePathAllowed(path string) bool {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	// Normalize path
	path = filepath.Clean(path)
	path = filepath.ToSlash(path)

	// Check if the path is explicitly allowed
	if _, exists := sc.allowedPaths[path]; exists {
		return true
	}

	// Check if the path is explicitly blocked
	for blocked := range sc.blockedPaths {
		if strings.Contains(path, blocked) {
			return false
		}
	}

	// Only trusted and admin can access files by default
	return sc.permissionLevel == PermissionLevelTrusted || sc.permissionLevel == PermissionLevelAdmin
}

// IsNetworkRequestAllowed checks if a network request is allowed
func (sc *SecurityChecker) IsNetworkRequestAllowed(urlStr, method string, permLevel PermissionLevel) bool {
	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Check node permission for HTTP requests
	if !sc.IsNodeAllowed("http-request", permLevel) {
		return false
	}

	// Check if the domain is allowed
	if !sc.IsDomainAllowed(parsedURL.Hostname()) {
		return false
	}

	// Additional checks for specific methods
	if method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH" {
		// More dangerous methods require higher permission
		return permLevel == PermissionLevelTrusted || permLevel == PermissionLevelAdmin
	}

	return true
}

// IsDataAccessAllowed checks if access to a data value is allowed
func (sc *SecurityChecker) IsDataAccessAllowed(dataID string, permLevel PermissionLevel) bool {
	// Check for sensitive data patterns
	sensitivePatterns := []string{
		"(?i)password",
		"(?i)secret",
		"(?i)token",
		"(?i)key",
		"(?i)credential",
		"(?i)auth",
	}

	for _, pattern := range sensitivePatterns {
		match, _ := regexp.MatchString(pattern, dataID)
		if match {
			// Only admin can access sensitive data
			return permLevel == PermissionLevelAdmin
		}
	}

	// Non-sensitive data is accessible to all permission levels
	return true
}

// IsCommandExecutionAllowed checks if executing a system command is allowed
func (sc *SecurityChecker) IsCommandExecutionAllowed(command string, permLevel PermissionLevel) bool {
	// Check node permission for command execution
	if !sc.IsNodeAllowed("execute-command", permLevel) {
		return false
	}

	// Only admin can execute commands
	if permLevel != PermissionLevelAdmin {
		return false
	}

	// Check for dangerous command patterns
	dangerousPatterns := []string{
		"rm -rf",
		"mkfs",
		"dd",
		"format",
		";",  // Command chaining
		"&&", // Command chaining
		"||", // Command chaining
		"|",  // Pipe
		">",  // Redirect
		"<",  // Redirect
		"sudo",
		"su ",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(command, pattern) {
			return false
		}
	}

	return true
}

// AllowDomain adds a domain to the allowed list
func (sc *SecurityChecker) AllowDomain(domain string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.allowedDomains[strings.ToLower(domain)] = true
}

// BlockDomain adds a domain to the blocked list
func (sc *SecurityChecker) BlockDomain(domain string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.blockedDomains[strings.ToLower(domain)] = true
}

// AllowPath adds a file path to the allowed list
func (sc *SecurityChecker) AllowPath(path string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.allowedPaths[filepath.ToSlash(filepath.Clean(path))] = true
}

// BlockPath adds a file path to the blocked list
func (sc *SecurityChecker) BlockPath(path string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.blockedPaths[filepath.ToSlash(filepath.Clean(path))] = true
}

// SetNodePermission sets permission for a specific node type
func (sc *SecurityChecker) SetNodePermission(nodeType string, permLevel PermissionLevel, allowed bool) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	if _, exists := sc.nodePermissions[nodeType]; !exists {
		sc.nodePermissions[nodeType] = make(map[PermissionLevel]bool)
	}
	sc.nodePermissions[nodeType][permLevel] = allowed
}

// ValidateContent checks if content contains potentially harmful data
func (sc *SecurityChecker) ValidateContent(content string) (bool, string) {
	// Check for potentially dangerous content patterns
	dangerousPatterns := []struct {
		pattern string
		reason  string
	}{
		{`(?i)<script`, "Contains script tags"},
		{`(?i)javascript:`, "Contains JavaScript URI"},
		{`(?i)data:text/html`, "Contains HTML data URI"},
		{`(?i)onerror=`, "Contains JavaScript event handler"},
		{`(?i)onclick=`, "Contains JavaScript event handler"},
		{`(?i)onload=`, "Contains JavaScript event handler"},
		{`(?i)eval\(`, "Contains JavaScript eval function"},
		{`(?i)document\.cookie`, "Attempts to access cookies"},
		{`(?i)localStorage`, "Attempts to access local storage"},
		{`(?i)sessionStorage`, "Attempts to access session storage"},
	}

	for _, dp := range dangerousPatterns {
		match, _ := regexp.MatchString(dp.pattern, content)
		if match {
			return false, fmt.Sprintf("Content validation failed: %s", dp.reason)
		}
	}

	return true, ""
}

// GetPermissionLevel returns the current permission level
func (sc *SecurityChecker) GetPermissionLevel() PermissionLevel {
	return sc.permissionLevel
}

// SetPermissionLevel updates the permission level
func (sc *SecurityChecker) SetPermissionLevel(level PermissionLevel) {
	sc.permissionLevel = level
}
