// Package scoring_test tests scoring logic.
package scoring_test

import (
	"math"
	"testing"

	"diffscope/pkg/analysis"
	"diffscope/pkg/config"
	"diffscope/pkg/scoring"
)

func TestScoreResult(t *testing.T) {
	cfg := config.DefaultConfig()
	s := scoring.NewScorer(cfg)

	result := &analysis.ImpactResult{
		Summary: analysis.Summary{
			TotalChanges: 100,
			HasBreaking:  true,
			HasTestFiles: false,
		},
		Files: []*analysis.FileImpact{
			{Path: "main.go", Additions: 50, Deletions: 50, HasBreaking: true, HasImports: true},
			{Path: "test.go", Additions: 10, Deletions: 0, TestFile: true},
		},
		Findings: []analysis.Finding{
			{Type: "security", Impact: analysis.Critical},
			{Type: "breaking", Impact: analysis.High},
		},
	}

	score := s.ScoreResult(result)
	if score.Overall < 0 || score.Overall > 100 {
		t.Errorf("score out of range: %d", score.Overall)
	}
}

func TestScoreEmpty(t *testing.T) {
	cfg := config.DefaultConfig()
	s := scoring.NewScorer(cfg)

	result := &analysis.ImpactResult{
		Summary:  analysis.Summary{TotalChanges: 0},
		Files:    []*analysis.FileImpact{},
		Findings: []analysis.Finding{},
	}

	score := s.ScoreResult(result)
	if score.Overall != 0 {
		t.Errorf("expected score 0 for empty diff, got %d", score.Overall)
	}
}

func TestSeverityLevels(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{0, "safe"},
		{10, "safe"},
		{14, "safe"},
		{15, "minor"},
		{25, "minor"},
		{34, "minor"},
		{35, "moderate"},
		{45, "moderate"},
		{54, "moderate"},
		{55, "significant"},
		{65, "significant"},
		{74, "significant"},
		{75, "critical"},
		{90, "critical"},
		{100, "critical"},
	}

	for _, tt := range tests {
		result := scoring.SeverityFromScoreTest(tt.score)
		if string(result) != tt.expected {
			t.Errorf("SeverityFromScore(%d) = %q, want %q", tt.score, result, tt.expected)
		}
	}
}

func TestMeetsThreshold(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Defaults.ScoringThreshold = 50
	s := scoring.NewScorer(cfg)

	score := &scoring.Score{Overall: 30}
	if !s.MeetsThreshold(score) {
		t.Error("expected threshold met for score 30")
	}

	score = &scoring.Score{Overall: 70}
	if s.MeetsThreshold(score) {
		t.Error("expected threshold not met for score 70")
	}
}

func TestSizeScore(t *testing.T) {
	cfg := config.DefaultConfig()
	s := scoring.NewScorer(cfg)

	result := &analysis.ImpactResult{
		Summary: analysis.Summary{TotalChanges: 0},
	}
	details := s.SizeScore(result)
	if details != 0 {
		t.Errorf("expected size score 0 for empty diff, got %f", details)
	}

	result = &analysis.ImpactResult{
		Summary: analysis.Summary{TotalChanges: 1000},
	}
	details = s.SizeScore(result)
	if details > 40 {
		t.Errorf("size score capped too low: %f", details)
	}
}

func TestBreakingScore(t *testing.T) {
	cfg := config.DefaultConfig()
	s := scoring.NewScorer(cfg)

	result := &analysis.ImpactResult{
		Files: []*analysis.FileImpact{
			{HasBreaking: true},
		},
		Findings: []analysis.Finding{
			{Type: "breaking"},
		},
	}
	score := s.BreakingScore(result)
	if score == 0 {
		t.Error("expected positive breaking score")
	}
}

func TestSecurityScore(t *testing.T) {
	cfg := config.DefaultConfig()
	s := scoring.NewScorer(cfg)

	result := &analysis.ImpactResult{
		Findings: []analysis.Finding{
			{Type: "security", Impact: analysis.Critical},
		},
	}
	score := s.SecurityScore(result)
	if score == 0 {
		t.Error("expected positive security score for critical finding")
	}
}

func TestTestImpactScore(t *testing.T) {
	cfg := config.DefaultConfig()
	s := scoring.NewScorer(cfg)

	// With test files — should be 0
	result := &analysis.ImpactResult{
		Summary: analysis.Summary{HasTestFiles: true},
	}
	score := s.TestImpactScore(result)
	if score != 0 {
		t.Errorf("expected test impact score 0 with test files, got %f", score)
	}

	// Without test files but with function changes
	result = &analysis.ImpactResult{
		Files: []*analysis.FileImpact{
			{Functions: []analysis.FuncInfo{{Name: "Foo"}, {Name: "Bar"}}},
		},
	}
	score = s.TestImpactScore(result)
	if score <= 0 {
		t.Errorf("expected positive test impact score, got %f", score)
	}
}

func TestWeightedScore(t *testing.T) {
	cfg := config.DefaultConfig()
	s := scoring.NewScorer(cfg)

	details := scoring.ScoreDetails{
		SizeScore:       20,
		BreakingScore:   15,
		SecurityScore:   10,
		TestImpactScore: 5,
		ImportScore:     3,
	}
	overall := s.WeightedScore(details)
	// 20*0.25 + 15*0.30 + 10*0.25 + 5*0.10 + 3*0.10 = 5 + 4.5 + 2.5 + 0.5 + 0.3 = 12.8
	expected := 12.8
	if math.Abs(overall-expected) > 0.01 {
		t.Errorf("weighted score: got %f, want %f", overall, expected)
	}
}
