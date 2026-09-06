package domain

// Severity levels.
const (
	SeverityLow      = "LOW"
	SeverityMedium   = "MEDIUM"
	SeverityHigh     = "HIGH"
	SeverityCritical = "CRITICAL"
)

// Status values.
const (
	StatusOpen          = "OPEN"
	StatusInvestigating = "INVESTIGATING"
	StatusSolved        = "SOLVED"
	StatusKnown         = "KNOWN"
)

// Environment values.
const (
	EnvLocal = "LOCAL"
	EnvDev   = "DEV"
	EnvUAT   = "UAT"
	EnvProd  = "PROD"
)

// Relation types for problem_related.
const (
	RelationRelated    = "RELATED"
	RelationSimilar    = "SIMILAR"
	RelationCausedBy   = "CAUSED_BY"
	RelationDuplicate  = "DUPLICATE"
	RelationWorkaround = "WORKAROUND"
)

// Severities is the ordered list of valid severities.
var Severities = []string{SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical}

// Statuses is the ordered list of valid statuses.
var Statuses = []string{StatusOpen, StatusInvestigating, StatusSolved, StatusKnown}

// Environments is the ordered list of valid environments.
var Environments = []string{EnvLocal, EnvDev, EnvUAT, EnvProd}

// RelationTypes is the ordered list of valid relation types.
var RelationTypes = []string{
	RelationRelated, RelationSimilar, RelationCausedBy, RelationDuplicate, RelationWorkaround,
}

// severityRank drives sorting by urgency.
var severityRank = map[string]int{
	SeverityLow: 1, SeverityMedium: 2, SeverityHigh: 3, SeverityCritical: 4,
}

func valid(v string, set []string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

// ValidSeverity reports whether v is an accepted severity.
func ValidSeverity(v string) bool { return valid(v, Severities) }

// ValidStatus reports whether v is an accepted status.
func ValidStatus(v string) bool { return valid(v, Statuses) }

// ValidEnvironment reports whether v is an accepted environment.
func ValidEnvironment(v string) bool { return valid(v, Environments) }

// ValidRelationType reports whether v is an accepted relation type.
func ValidRelationType(v string) bool { return valid(v, RelationTypes) }

// SeverityRank returns a sortable weight for a severity (unknown -> 0).
func SeverityRank(v string) int { return severityRank[v] }
