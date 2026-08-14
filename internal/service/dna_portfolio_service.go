package service

import (
	"sort"

	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

func buildDNAPortfolio(rows []models.DNAPortfolioRow) models.DNAPortfolioResponse {
	result := models.DNAPortfolioResponse{Summary: models.DNAPortfolioSummary{ByTier: map[string]int{}, ByIndustry: map[string]int{}}, Items: make([]models.DNAPortfolioItem, 0)}
	for _, row := range rows {
		if row.DNA == nil {
			continue
		}
		result.Summary.TotalProfiles++
		if row.Tier != nil {
			result.Summary.ByTier[*row.Tier]++
		}
		if row.Industry != nil {
			result.Summary.ByIndustry[*row.Industry]++
		}
		terms := make([]string, 0)
		seen := map[string]struct{}{}
		for _, term := range row.DNA.Vocabulary.Terms {
			if _, ok := seen[term.Term]; !ok {
				seen[term.Term] = struct{}{}
				terms = append(terms, term.Term)
			}
		}
		item := models.DNAPortfolioItem{AccountID: row.AccountID, Name: row.Name, Industry: row.Industry, Tier: row.Tier, ActiveSignalCount: row.ActiveSignalCount, Tone: row.DNA.Tone.Primary, Vocabulary: terms, ProblemFraming: row.DNA.ProblemFraming.Description, ProofStyle: row.DNA.ProofStyle.Primary, CTAStyle: row.DNA.CTAPatterns.Style, DoRules: nonNilStrings(row.DNA.DoRules), DontRules: nonNilStrings(row.DNA.DontRules), SignalTypes: nonNilStrings(row.SignalTypes)}
		sort.Strings(item.SignalTypes)
		result.Items = append(result.Items, item)
	}
	sort.SliceStable(result.Items, func(i, j int) bool {
		a, b := result.Items[i], result.Items[j]
		rank := func(t *string) int {
			if t == nil {
				return 5
			}
			switch *t {
			case "Focus Accounts":
				return 1
			case "Tier 1":
				return 2
			case "Tier 2":
				return 3
			case "Below ICP":
				return 4
			}
			return 5
		}
		if rank(a.Tier) != rank(b.Tier) {
			return rank(a.Tier) < rank(b.Tier)
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.AccountID.String() < b.AccountID.String()
	})
	return result
}

func compareDNAPortfolio(rows []models.DNAPortfolioRow) models.DNACompareResponse {
	result := models.DNACompareResponse{SelectedCount: len(rows), DominantTone: []models.DNAValueCount{}, SharedVocabulary: []models.DNAValueCount{}, UniqueVocabulary: []models.DNAValueCount{}, ProofStyles: []models.DNAValueCount{}, CTAStyles: []models.DNAValueCount{}, DoRules: []models.DNAValueCount{}, DontRules: []models.DNAValueCount{}, SignalTypes: []models.DNAValueCount{}, ProblemFraming: []models.DNAProblemFramingItem{}}
	count := func(values [][]string) []models.DNAValueCount {
		m := map[string]int{}
		for _, list := range values {
			seen := map[string]struct{}{}
			for _, v := range list {
				if v != "" {
					if _, ok := seen[v]; !ok {
						seen[v] = struct{}{}
						m[v]++
					}
				}
			}
		}
		out := make([]models.DNAValueCount, 0, len(m))
		for v, c := range m {
			out = append(out, models.DNAValueCount{Value: v, Count: c})
		}
		sort.Slice(out, func(i, j int) bool {
			if out[i].Count != out[j].Count {
				return out[i].Count > out[j].Count
			}
			return out[i].Value < out[j].Value
		})
		return out
	}
	tones, proofs, ctas := [][]string{}, [][]string{}, [][]string{}
	vocab, dos, donts, signals := [][]string{}, [][]string{}, [][]string{}, [][]string{}
	for _, row := range rows {
		dna := row.DNA
		if dna.Tone.Primary != nil {
			tones = append(tones, []string{*dna.Tone.Primary})
		}
		if dna.ProofStyle.Primary != nil {
			proofs = append(proofs, []string{*dna.ProofStyle.Primary})
		}
		if dna.CTAPatterns.Style != nil {
			ctas = append(ctas, []string{*dna.CTAPatterns.Style})
		}
		vocab = append(vocab, vocabularyTerms(dna))
		dos = append(dos, dna.DoRules)
		donts = append(donts, dna.DontRules)
		signals = append(signals, row.SignalTypes)
		result.ProblemFraming = append(result.ProblemFraming, models.DNAProblemFramingItem{AccountID: row.AccountID, AccountName: row.Name, Value: dna.ProblemFraming.Description})
	}
	result.DominantTone = count(tones)
	result.ProofStyles = count(proofs)
	result.CTAStyles = count(ctas)
	result.DoRules = count(dos)
	result.DontRules = count(donts)
	result.SignalTypes = count(signals)
	all := count(vocab)
	for _, v := range all {
		if v.Count >= 2 {
			result.SharedVocabulary = append(result.SharedVocabulary, v)
		} else {
			result.UniqueVocabulary = append(result.UniqueVocabulary, v)
		}
	}
	sort.Slice(result.ProblemFraming, func(i, j int) bool {
		if result.ProblemFraming[i].AccountName != result.ProblemFraming[j].AccountName {
			return result.ProblemFraming[i].AccountName < result.ProblemFraming[j].AccountName
		}
		return result.ProblemFraming[i].AccountID.String() < result.ProblemFraming[j].AccountID.String()
	})
	return result
}
func vocabularyTerms(dna *models.CommunicationDNA) []string {
	out := []string{}
	seen := map[string]struct{}{}
	for _, t := range dna.Vocabulary.Terms {
		if _, ok := seen[t.Term]; !ok {
			seen[t.Term] = struct{}{}
			out = append(out, t.Term)
		}
	}
	return out
}
func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
