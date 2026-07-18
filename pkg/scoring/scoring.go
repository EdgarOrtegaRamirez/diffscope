// Package scoring computes impact scores for diffs.
package scoring

import (
	"diffscope/pkg/analysis"
	"diffscope/pkg/config"
	"math"
)

// Score represents the overall impact score (0-100).
type Score struct {
	Overall  int           `json:"overall"`
	Severity SeverityLevel `json:"severity"`
	Details  ScoreDetails  `json:"details"`
}

// SeverityLevel represents the severity category.
type SeverityLevel string

const (
	Safe        SeverityLevel = "safe"
	Minor       SeverityLevel = "minor"
	Moderate    SeverityLevel = "moderate"
	Significant SeverityLevel = "significant"
	Critical    SeverityLevel = "critical"
)

// ScoreDetails contains the breakdown of the score.
type ScoreDetails struct {
	SizeScore       float64 `json:"size_score"`
	BreakingScore   float64 `json:"breaking_score"`
	SecurityScore   float64 `json:"security_score"`
	TestImpactScore float64 `json:"test_impact_score"`
	ImportScore     float64 `json:"import_score"`
}

// Scorer computes impact scores.
type Scorer struct {
	config config.Config
}

// NewScorer creates a new Scorer.
func NewScorer(cfg config.Config) *Scorer {
	return &Scorer{config: cfg}
}

// ScoreResult computes the impact score for an analysis result.
func (s *Scorer) ScoreResult(result *analysis.ImpactResult) *Score {
	details := ScoreDetails{}

	// Size score: based on total lines changed
	details.SizeScore = s.calculateSizeScore(result)

	// Breaking change score
	details.BreakingScore = s.calculateBreakingScore(result)

	// Security score
	details.SecurityScore = s.calculateSecurityScore(result)

	// Test impact score
	details.TestImpactScore = s.calculateTestImpactScore(result)

	// Import score
	details.ImportScore = s.calculateImportScore(result)

	// Weighted overall score
	overall := s.weightedScore(details)

	return &Score{
		Overall:  int(math.Round(overall)),
		Severity: severityFromScore(int(math.Round(overall))),
		Details:  details,
	}
}

// calculateSizeScore computes a score based on the number of lines changed.
func (s *Scorer) calculateSizeScore(result *analysis.ImpactResult) float64 {
	total := result.Summary.TotalChanges
	if total == 0 {
		return 0
	}

	// Logarithmic scaling: more changes = higher score, but diminishing returns
	score := math.Log10(float64(total)+1) * 20
	if score > 40 {
		score = 40
	}
	return score
}

// calculateBreakingScore computes a score based on breaking changes.
func (s *Scorer) calculateBreakingScore(result *analysis.ImpactResult) float64 {
	score := 0.0

	for _, f := range result.Files {
		if f.HasBreaking {
			score += 25
		}
	}
	for _, f := range result.Findings {
		if f.Type == "breaking" {
			score += 10
		}
	}

	if score > 30 {
		score = 30
	}
	return score
}

// calculateSecurityScore computes a score based on security findings.
func (s *Scorer) calculateSecurityScore(result *analysis.ImpactResult) float64 {
	score := 0.0

	for _, f := range result.Findings {
		if f.Type == "security" {
			switch f.Impact {
			case analysis.Critical:
				score += 20
			case analysis.High:
				score += 15
			case analysis.Medium:
				score += 10
			case analysis.Low:
				score += 5
			case analysis.Info:
				score += 2
			}
		}
	}

	if score > 30 {
		score = 30
	}
	return score
}

// calculateTestImpactScore computes a score based on test file impact.
func (s *Scorer) calculateTestImpactScore(result *analysis.ImpactResult) float64 {
	if result.Summary.HasTestFiles {
		// If there are test file changes, lower the score
		return 0
	}

	// If there are function changes but no test changes, higher score
	totalFuncs := 0
	for _, f := range result.Files {
		totalFuncs += len(f.Functions)
	}

	if totalFuncs > 0 {
		score := float64(totalFuncs) * 3
		if score > 20 {
			score = 20
		}
		return score
	}
	return 0
}

// calculateImportScore computes a score based on import changes.
func (s *Scorer) calculateImportScore(result *analysis.ImpactResult) float64 {
	score := 0.0
	for _, f := range result.Files {
		if f.HasImports {
			score += 5
		}
	}
	if score > 10 {
		score = 10
	}
	return score
}

// weightedScore combines all sub-scores into a single 0-100 score.
func (s *Scorer) weightedScore(details ScoreDetails) float64 {
	// Weights: size 25%, breaking 30%, security 25%, test impact 10%, imports 10%
	return details.SizeScore*0.25 +
		details.BreakingScore*0.30 +
		details.SecurityScore*0.25 +
		details.TestImpactScore*0.10 +
		details.ImportScore*0.10
}

// severityFromScore maps a numeric score to a severity level.
func severityFromScore(score int) SeverityLevel {
	switch {
	case score < 15:
		return Safe
	case score < 35:
		return Minor
	case score < 55:
		return Moderate
	case score < 75:
		return Significant
	default:
		return Critical
	}
}

// MeetsThreshold checks if the score meets the configured threshold.
func (s *Scorer) MeetsThreshold(score *Score) bool {
	return score.Overall < s.config.Defaults.ScoringThreshold
}

// Exported test helpers
func (s *Scorer) SizeScore(result *analysis.ImpactResult) float64 {
	return s.calculateSizeScore(result)
}
func (s *Scorer) BreakingScore(result *analysis.ImpactResult) float64 {
	return s.calculateBreakingScore(result)
}
func (s *Scorer) SecurityScore(result *analysis.ImpactResult) float64 {
	return s.calculateSecurityScore(result)
}
func (s *Scorer) TestImpactScore(result *analysis.ImpactResult) float64 {
	return s.calculateTestImpactScore(result)
}
func (s *Scorer) WeightedScore(details ScoreDetails) float64 {
	return s.weightedScore(details)
}

// SeverityFromScoreTest is an exported wrapper for testing.
func SeverityFromScoreTest(score int) SeverityLevel {
	return severityFromScore(score)
}
