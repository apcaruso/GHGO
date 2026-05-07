//go:build ghgo_devtools

package factors

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/ports"
)

type ImportOptions struct {
	Force bool
}

type ImportSummary struct {
	FactorSetID  string
	RowsRead     int
	RowsImported int
	RowsSkipped  int
	Warnings     []string
}

func ImportDEFRA2025(ctx context.Context, st ports.Store, path string, opts ImportOptions) (ImportSummary, error) {
	var summary ImportSummary
	if st == nil {
		return summary, fmt.Errorf("store is required")
	}
	if path == "" {
		return summary, fmt.Errorf("DEFRA 2025 workbook path is required")
	}

	existing, err := st.FindFactorSetBySourceYearVersion(ctx, defra2025Source, defra2025Year, defra2025Version)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		return summary, err
	}
	if existing != nil {
		summary.FactorSetID = existing.ID
		count, err := st.CountEmissionFactorsBySet(ctx, existing.ID)
		if err != nil {
			return summary, err
		}
		if count > 0 && !opts.Force {
			return summary, nil
		}
	}

	parsed, err := parseDEFRA2025File(path)
	if err != nil {
		return summary, err
	}
	summary.RowsRead = parsed.RowsRead
	summary.RowsSkipped = parsed.RowsSkipped
	summary.Warnings = parsed.Warnings

	if err := st.WithTx(ctx, func(tx ports.Store) error {
		factorSet := existing
		if factorSet == nil {
			factorSet = &domain.FactorSet{
				ID:           defra2025FactorSetID,
				Name:         defra2025Name,
				Source:       defra2025Source,
				Year:         defra2025Year,
				Version:      defra2025Version,
				ImportedAt:   time.Now().UTC(),
				MetadataJSON: `{}`,
			}
			if err := tx.CreateFactorSet(ctx, *factorSet); err != nil {
				return err
			}
		}

		if opts.Force && existing != nil {
			if err := tx.DeleteEmissionFactorsBySet(ctx, factorSet.ID); err != nil {
				return err
			}
		}

		for _, candidate := range parsed.Candidates {
			id, err := newID("emission_factor")
			if err != nil {
				return err
			}
			if err := tx.CreateEmissionFactor(ctx, candidate.emissionFactor(id, factorSet.ID)); err != nil {
				return err
			}
		}

		summary.FactorSetID = factorSet.ID
		summary.RowsImported = len(parsed.Candidates)
		return nil
	}); err != nil {
		return summary, err
	}

	return summary, nil
}

func newID(prefix string) (domain.ID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return domain.ID(prefix + "_" + hex.EncodeToString(b[:])), nil
}
