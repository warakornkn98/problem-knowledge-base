package domain

import (
	"testing"
	"time"
)

func TestValidators(t *testing.T) {
	if !ValidSeverity(SeverityCritical) || ValidSeverity("URGENT") {
		t.Fatal("severity validation wrong")
	}
	if !ValidStatus(StatusSolved) || ValidStatus("DONE") {
		t.Fatal("status validation wrong")
	}
	if !ValidEnvironment(EnvProd) || ValidEnvironment("STAGING") {
		t.Fatal("environment validation wrong")
	}
	if !ValidRelationType(RelationDuplicate) || ValidRelationType("COPY") {
		t.Fatal("relation type validation wrong")
	}
}

func TestSeverityRankOrdering(t *testing.T) {
	if !(SeverityRank(SeverityLow) < SeverityRank(SeverityMedium) &&
		SeverityRank(SeverityMedium) < SeverityRank(SeverityHigh) &&
		SeverityRank(SeverityHigh) < SeverityRank(SeverityCritical)) {
		t.Fatal("severity rank not monotonically increasing")
	}
	if SeverityRank("nonsense") != 0 {
		t.Fatal("unknown severity should rank 0")
	}
}

func TestProblemValidate(t *testing.T) {
	p := &Problem{Severity: SeverityHigh, Status: StatusOpen, Environment: EnvProd}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
	p.Status = "BOGUS"
	if err := p.Validate(); err == nil {
		t.Fatal("expected invalid enum error")
	}
}

func TestApplyStatusTimestamps(t *testing.T) {
	now := time.Now()
	p := &Problem{Status: StatusSolved}
	p.ApplyStatusTimestamps(now)
	if p.SolvedAt == nil {
		t.Fatal("solved problem should get solved_at")
	}
	p.Status = StatusOpen
	p.ApplyStatusTimestamps(now)
	if p.SolvedAt != nil {
		t.Fatal("re-opened problem should clear solved_at")
	}
}
