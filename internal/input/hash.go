package input

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"ghgo/internal/domain"
)

type hashPart struct {
	name  string
	value string
}

func RawHash(raw string) string {
	return sha256Hex(normalizeLineEndings(raw))
}

func SourceHash(record domain.ActivityRecord) string {
	parts := []hashPart{
		{name: "organization_id", value: record.OrganizationID},
		{name: "reporting_period_id", value: record.ReportingPeriodID},
	}

	switch record.SourceKind {
	case domain.ActivitySourceKindElectricityMonthlyKWh, domain.ActivitySourceKindNaturalGasMonthlySMC:
		parts = append(parts,
			hashPart{name: "facility_id", value: facilityHashValue(record.FacilityID)},
			hashPart{name: "source_kind", value: string(record.SourceKind)},
			hashPart{name: "period_start", value: hashTime(record.PeriodStart)},
			hashPart{name: "period_end", value: hashTime(record.PeriodEnd)},
			hashPart{name: "amount", value: hashAmount(record.Amount)},
			hashPart{name: "unit", value: record.Unit},
		)
	case domain.ActivitySourceKindMobileFuelLitres:
		parts = appendOptionalFacility(parts, record.FacilityID)
		parts = append(parts,
			hashPart{name: "source_kind", value: string(record.SourceKind)},
			hashPart{name: "fuel_type", value: record.FuelType},
			hashPart{name: "amount", value: hashAmount(record.Amount)},
			hashPart{name: "unit", value: record.Unit},
		)
	case domain.ActivitySourceKindVehicleDistanceKM:
		parts = appendOptionalFacility(parts, record.FacilityID)
		parts = append(parts,
			hashPart{name: "source_kind", value: string(record.SourceKind)},
			hashPart{name: "vehicle_name", value: record.VehicleName},
			hashPart{name: "plate", value: record.Plate},
			hashPart{name: "vehicle_type", value: record.VehicleType},
			hashPart{name: "vehicle_size_class", value: record.VehicleSizeClass},
			hashPart{name: "fuel_type", value: record.FuelType},
			hashPart{name: "amount", value: hashAmount(record.Amount)},
			hashPart{name: "unit", value: record.Unit},
		)
	case domain.ActivitySourceKindRefrigerantsAnnualKG:
		parts = append(parts,
			hashPart{name: "facility_id", value: facilityHashValue(record.FacilityID)},
			hashPart{name: "source_kind", value: string(record.SourceKind)},
			hashPart{name: "substance", value: record.Substance},
			hashPart{name: "amount", value: hashAmount(record.Amount)},
			hashPart{name: "unit", value: record.Unit},
		)
	}

	return hashParts(parts)
}

func normalizeLineEndings(raw string) string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	return strings.ReplaceAll(raw, "\r", "\n")
}

func appendOptionalFacility(parts []hashPart, facilityID *domain.ID) []hashPart {
	if facilityID == nil {
		return parts
	}
	return append(parts, hashPart{name: "facility_id", value: *facilityID})
}

func facilityHashValue(facilityID *domain.ID) string {
	if facilityID == nil {
		return ""
	}
	return *facilityID
}

func hashParts(parts []hashPart) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.name)
		b.WriteByte('=')
		b.WriteString(part.value)
		b.WriteByte('\n')
	}
	return sha256Hex(b.String())
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func hashTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func hashAmount(amount float64) string {
	return strconv.FormatFloat(amount, 'f', -1, 64)
}
