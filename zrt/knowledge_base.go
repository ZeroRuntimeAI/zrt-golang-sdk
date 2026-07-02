package zrt

import "cmp"

// KnowledgeBaseConfig configures a RAG knowledge base.
type KnowledgeBaseConfig struct {
	// Provider is the RAG backend id. Defaults to "custom".
	Provider string // default "custom"
	// IndexName is the index to query; reconciled with ID.
	IndexName string
	// TopK is the number of results to retrieve. Defaults to 5.
	TopK int // default 5
	// MinScore is the minimum relevance score for a result. Defaults to 0.7.
	MinScore float64 // default 0.7
	// Params holds provider-specific parameters.
	Params map[string]string
	// ID is the knowledge base identifier; reconciled with IndexName.
	ID string
}

// normalize applies default values and reconciles the index_name/id fields.
func (cfg *KnowledgeBaseConfig) normalize() {
	cfg.Provider = cmp.Or(cfg.Provider, "custom")
	cfg.TopK = cmp.Or(cfg.TopK, 5)
	cfg.MinScore = cmp.Or(cfg.MinScore, 0.7)
	// Reconcile index_name <-> id.
	switch {
	case cfg.IndexName != "" && cfg.ID != "" && cfg.IndexName != cfg.ID:
		cfg.ID = cfg.IndexName
	case cfg.ID != "" && cfg.IndexName == "":
		cfg.IndexName = cfg.ID
	case cfg.IndexName != "" && cfg.ID == "":
		cfg.ID = cfg.IndexName
	}
}

// KnowledgeBase is a RAG knowledge base descriptor.
type KnowledgeBase struct {
	// Config is the normalized knowledge base configuration.
	Config *KnowledgeBaseConfig
}

// NewKnowledgeBase builds a KnowledgeBase from config, applying defaults.
// Pass a zero &KnowledgeBaseConfig{} (or nil) to accept all defaults.
func NewKnowledgeBase(config *KnowledgeBaseConfig) *KnowledgeBase {
	if config == nil {
		config = &KnowledgeBaseConfig{}
	}
	config.normalize()
	return &KnowledgeBase{Config: config}
}
