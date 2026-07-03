package history

type Decision string

const (
	DecisionCandidate Decision = "candidate"
	DecisionReject    Decision = "reject"
)

type ValidateStatus string

const (
	ValidatePassed ValidateStatus = "passed"
	ValidateFailed ValidateStatus = "failed"
)

type Record struct {
	ID                    string
	RawSQLHash            string
	NormalizedFingerprint string
	ParamsHash            string
	RuleID                string
	RuleVersion           int
	TemplateChecksum      string
	SchemaSignature       string
	RewrittenSQL          string
	Decision              Decision
	ValidateStatus        ValidateStatus
	RuleEnabled           bool
}

type Query struct {
	RawSQLHash            string
	NormalizedFingerprint string
	ParamsHash            string
	RuleID                string
	RuleVersion           int
	TemplateChecksum      string
	SchemaSignature       string
}

type MemoryStore struct {
	records []Record
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Save(record Record) {
	s.records = append(s.records, record)
}

func (s *MemoryStore) LookupExact(query Query) (Record, bool) {
	if s == nil || query.RawSQLHash == "" {
		return Record{}, false
	}
	for i := len(s.records) - 1; i >= 0; i-- {
		record := s.records[i]
		if record.RawSQLHash == query.RawSQLHash && reusable(record, query) {
			return record, true
		}
	}
	return Record{}, false
}

func (s *MemoryStore) LookupFingerprint(query Query) (Record, bool) {
	if s == nil || query.NormalizedFingerprint == "" {
		return Record{}, false
	}
	for i := len(s.records) - 1; i >= 0; i-- {
		record := s.records[i]
		if record.NormalizedFingerprint == query.NormalizedFingerprint &&
			record.ParamsHash == query.ParamsHash &&
			reusable(record, query) {
			return record, true
		}
	}
	return Record{}, false
}

func reusable(record Record, query Query) bool {
	return record.Decision == DecisionCandidate &&
		record.ValidateStatus == ValidatePassed &&
		record.RuleEnabled &&
		record.RuleID == query.RuleID &&
		query.RuleID != "" &&
		record.ParamsHash == query.ParamsHash &&
		query.ParamsHash != "" &&
		record.RuleVersion == query.RuleVersion &&
		record.TemplateChecksum == query.TemplateChecksum &&
		record.SchemaSignature == query.SchemaSignature
}
