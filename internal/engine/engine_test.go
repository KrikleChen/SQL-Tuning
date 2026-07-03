package engine

import (
	"testing"

	"github.com/KrikleChen/SQL-Tuning/internal/history"
)

func TestRewriteUsesExactHistoryBeforeRules(t *testing.T) {
	store := history.NewMemoryStore()
	store.Save(history.Record{
		ID:                    "hist_001",
		RawSQLHash:            HashRawSQL("select * from slow"),
		NormalizedFingerprint: FingerprintSQL("select * from slow"),
		ParamsHash:            "params:none",
		RuleID:                "rule",
		RuleVersion:           1,
		TemplateChecksum:      "tmpl",
		SchemaSignature:       "schema",
		RewrittenSQL:          "select * from fast",
		Decision:              history.DecisionCandidate,
		ValidateStatus:        history.ValidatePassed,
		RuleEnabled:           true,
	})

	e := New(store, Rule{
		ID:               "rule",
		Version:          1,
		TemplateChecksum: "tmpl",
		SchemaSignature:  "schema",
		RawSQL:           "select * from slow",
		RewrittenSQL:     "select * from rule",
	})

	got := e.Rewrite(Request{SQL: "select * from slow"})
	if got.Source != SourceHistory || got.SQL != "select * from fast" {
		t.Fatalf("expected history rewrite, got %#v", got)
	}
}

func TestRewriteUsesFingerprintHistoryBeforeRules(t *testing.T) {
	store := history.NewMemoryStore()
	store.Save(history.Record{
		ID:                    "hist_001",
		RawSQLHash:            HashRawSQL("select * from slow"),
		NormalizedFingerprint: FingerprintSQL("select * from slow"),
		ParamsHash:            "params:none",
		RuleID:                "rule",
		RuleVersion:           1,
		TemplateChecksum:      "tmpl",
		SchemaSignature:       "schema",
		RewrittenSQL:          "select * from fast",
		Decision:              history.DecisionCandidate,
		ValidateStatus:        history.ValidatePassed,
		RuleEnabled:           true,
	})

	e := New(store, Rule{
		ID:               "rule",
		Version:          1,
		TemplateChecksum: "tmpl",
		SchemaSignature:  "schema",
		RawSQL:           "select * from slow",
		RewrittenSQL:     "select * from rule",
	})

	got := e.Rewrite(Request{SQL: " select  *\nfrom   slow "})
	if got.Source != SourceHistory || got.SQL != "select * from fast" {
		t.Fatalf("expected fingerprint history rewrite, got %#v", got)
	}
}

func TestRewriteInvalidatesHistoryWhenRuleVersionChanges(t *testing.T) {
	store := history.NewMemoryStore()
	store.Save(history.Record{
		ID:                    "hist_001",
		RawSQLHash:            HashRawSQL("select * from slow"),
		NormalizedFingerprint: FingerprintSQL("select * from slow"),
		ParamsHash:            "params:none",
		RuleID:                "rule",
		RuleVersion:           1,
		TemplateChecksum:      "tmpl",
		SchemaSignature:       "schema",
		RewrittenSQL:          "select * from stale_history",
		Decision:              history.DecisionCandidate,
		ValidateStatus:        history.ValidatePassed,
		RuleEnabled:           true,
	})

	e := New(store, Rule{
		ID:               "rule",
		Version:          2,
		TemplateChecksum: "tmpl",
		SchemaSignature:  "schema",
		RawSQL:           "select * from slow",
		RewrittenSQL:     "select * from current_rule",
	})

	got := e.Rewrite(Request{SQL: "select * from slow"})
	if got.Source != SourceRule || got.SQL != "select * from current_rule" {
		t.Fatalf("expected stale history to be skipped, got %#v", got)
	}
}

func TestRewriteUsesRuleWhenNoHistoryMatches(t *testing.T) {
	e := New(history.NewMemoryStore(), Rule{
		ID:               "rule",
		Version:          1,
		TemplateChecksum: "tmpl",
		SchemaSignature:  "schema",
		RawSQL:           "select * from slow",
		RewrittenSQL:     "select * from rule",
	})

	got := e.Rewrite(Request{SQL: " select  *\nfrom   slow "})
	if got.Source != SourceRule || got.SQL != "select * from rule" {
		t.Fatalf("expected rule rewrite, got %#v", got)
	}
}

func TestRewriteFallsBackToOriginalWhenNoRuleMatches(t *testing.T) {
	e := New(history.NewMemoryStore())
	got := e.Rewrite(Request{SQL: "select * from unknown"})
	if got.Source != SourceOriginal || got.SQL != "select * from unknown" {
		t.Fatalf("expected original fallback, got %#v", got)
	}
}
