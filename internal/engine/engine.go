package engine

import (
	"github.com/KrikleChen/SQL-Tuning/internal/fingerprint"
	"github.com/KrikleChen/SQL-Tuning/internal/history"
	"github.com/KrikleChen/SQL-Tuning/internal/normalize"
)

const defaultParamsHash = "params:none"

type Source string

const (
	SourceHistory  Source = "history"
	SourceRule     Source = "rule"
	SourceOriginal Source = "original"
)

type Request struct {
	SQL string
}

type Result struct {
	SQL    string
	Source Source
	Reason string
}

type Rule struct {
	ID               string
	Version          int
	TemplateChecksum string
	SchemaSignature  string
	RawSQL           string
	RewrittenSQL     string
}

type Engine struct {
	history *history.MemoryStore
	rules   []Rule
}

func New(store *history.MemoryStore, rules ...Rule) *Engine {
	if store == nil {
		store = history.NewMemoryStore()
	}
	return &Engine{history: store, rules: rules}
}

func (e *Engine) Rewrite(request Request) Result {
	if e == nil {
		return Result{SQL: request.SQL, Source: SourceOriginal, Reason: "no_engine"}
	}

	rawHash := HashRawSQL(request.SQL)
	normalizedFingerprint := FingerprintSQL(request.SQL)

	for _, rule := range e.rules {
		query := queryForRule(rule, rawHash, normalizedFingerprint)
		if record, ok := e.history.LookupExact(query); ok {
			return Result{SQL: record.RewrittenSQL, Source: SourceHistory, Reason: "history_exact_hit"}
		}
	}

	for _, rule := range e.rules {
		query := queryForRule(rule, rawHash, normalizedFingerprint)
		if record, ok := e.history.LookupFingerprint(query); ok {
			return Result{SQL: record.RewrittenSQL, Source: SourceHistory, Reason: "history_fingerprint_hit"}
		}
	}

	for _, rule := range e.rules {
		if FingerprintSQL(rule.RawSQL) == normalizedFingerprint {
			return Result{SQL: rule.RewrittenSQL, Source: SourceRule, Reason: "rule_match"}
		}
	}

	return Result{SQL: request.SQL, Source: SourceOriginal, Reason: "no_match"}
}

func queryForRule(rule Rule, rawHash, normalizedFingerprint string) history.Query {
	return history.Query{
		RawSQLHash:            rawHash,
		NormalizedFingerprint: normalizedFingerprint,
		ParamsHash:            defaultParamsHash,
		RuleID:                rule.ID,
		RuleVersion:           rule.Version,
		TemplateChecksum:      rule.TemplateChecksum,
		SchemaSignature:       rule.SchemaSignature,
	}
}

func HashRawSQL(sql string) string {
	return fingerprint.SHA256(sql)
}

func FingerprintSQL(sql string) string {
	return fingerprint.SHA256(normalize.SQL(sql))
}
