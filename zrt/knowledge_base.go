package zrt

// KnowledgeBaseConfig configures a RAG knowledge base.
type KnowledgeBaseConfig struct {
	Provider  string // default "custom"
	IndexName string
	TopK      int     // default 5
	MinScore  float64 // default 0.7
	Params    map[string]string
	ID        string
}

// normalize applies default values and reconciles the index_name/id fields.
func (cfg *KnowledgeBaseConfig) normalize() {
	if cfg.Provider == "" {
		cfg.Provider = "custom"
	}
	if cfg.TopK == 0 {
		cfg.TopK = 5
	}
	if cfg.MinScore == 0 {
		cfg.MinScore = 0.7
	}
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
