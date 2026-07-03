package history

import "testing"

func reusableRecord() Record {
	return Record{
		ID:                    "hist_001",
		RawSQLHash:            "raw",
		NormalizedFingerprint: "fp",
		ParamsHash:            "params:none",
		RuleID:                "rule",
		RuleVersion:           1,
		TemplateChecksum:      "tmpl",
		SchemaSignature:       "schema",
		RewrittenSQL:          "select 1",
		Decision:              DecisionCandidate,
		ValidateStatus:        ValidatePassed,
		RuleEnabled:           true,
	}
}

func reusableQuery() Query {
	return Query{
		RawSQLHash:            "raw",
		NormalizedFingerprint: "fp",
		ParamsHash:            "params:none",
		RuleID:                "rule",
		RuleVersion:           1,
		TemplateChecksum:      "tmpl",
		SchemaSignature:       "schema",
	}
}

func TestLookupExactReturnsReusableCandidate(t *testing.T) {
	store := NewMemoryStore()
	record := reusableRecord()
	store.Save(record)

	got, ok := store.LookupExact(reusableQuery())
	if !ok || got.RewrittenSQL != "select 1" {
		t.Fatalf("expected reusable history hit, got %#v ok=%v", got, ok)
	}
}

func TestLookupFingerprintReturnsReusableCandidate(t *testing.T) {
	store := NewMemoryStore()
	record := reusableRecord()
	record.RawSQLHash = "different raw"
	store.Save(record)

	got, ok := store.LookupFingerprint(reusableQuery())
	if !ok || got.ID != "hist_001" {
		t.Fatalf("expected fingerprint history hit, got %#v ok=%v", got, ok)
	}
}

func TestLookupRejectsNonCandidate(t *testing.T) {
	store := NewMemoryStore()
	record := reusableRecord()
	record.Decision = DecisionReject
	store.Save(record)

	if got, ok := store.LookupExact(reusableQuery()); ok {
		t.Fatalf("expected non-candidate miss, got %#v", got)
	}
}

func TestLookupRejectsUnvalidatedCandidate(t *testing.T) {
	store := NewMemoryStore()
	record := reusableRecord()
	record.ValidateStatus = ValidateFailed
	store.Save(record)

	if got, ok := store.LookupExact(reusableQuery()); ok {
		t.Fatalf("expected unvalidated candidate miss, got %#v", got)
	}
}

func TestLookupRejectsMissingRuleIDOrParamsHash(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Query)
	}{
		{
			name: "missing rule id",
			mutate: func(query *Query) {
				query.RuleID = ""
			},
		},
		{
			name: "missing params hash",
			mutate: func(query *Query) {
				query.ParamsHash = ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore()
			store.Save(reusableRecord())
			query := reusableQuery()
			tt.mutate(&query)

			if got, ok := store.LookupFingerprint(query); ok {
				t.Fatalf("expected miss when %s, got %#v", tt.name, got)
			}
		})
	}
}

func TestLookupInvalidatesWhenRuleVersionTemplateOrSchemaChanges(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Query)
	}{
		{
			name: "rule version",
			mutate: func(query *Query) {
				query.RuleVersion = 2
			},
		},
		{
			name: "template checksum",
			mutate: func(query *Query) {
				query.TemplateChecksum = "new-template"
			},
		},
		{
			name: "schema signature",
			mutate: func(query *Query) {
				query.SchemaSignature = "new-schema"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore()
			store.Save(reusableRecord())
			query := reusableQuery()
			tt.mutate(&query)

			if got, ok := store.LookupExact(query); ok {
				t.Fatalf("expected miss after %s changed, got %#v", tt.name, got)
			}
		})
	}
}
