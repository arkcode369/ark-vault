package domain

import "strings"

// AssetType represents the category of a traded instrument.
type AssetType string

const (
	AssetForex   AssetType = "FOREX"
	AssetGold    AssetType = "GOLD"
	AssetIndices AssetType = "INDICES"
	AssetCrypto  AssetType = "CRYPTO"
)

// ValidAssetTypes lists all recognised asset types.
var ValidAssetTypes = []AssetType{AssetForex, AssetGold, AssetIndices, AssetCrypto}

// IsValid returns true when the asset type is one of the known constants.
func (a AssetType) IsValid() bool {
	for _, v := range ValidAssetTypes {
		if a == v {
			return true
		}
	}
	return false
}

// String returns a human-friendly label.
func (a AssetType) String() string {
	switch a {
	case AssetForex:
		return "Forex"
	case AssetGold:
		return "Gold"
	case AssetIndices:
		return "Indices"
	case AssetCrypto:
		return "Crypto"
	default:
		return string(a)
	}
}

// knownSymbols maps common pairs/symbols → asset type.
var knownSymbols = map[string]AssetType{
	// Gold
	"XAUUSD": AssetGold, "GOLD": AssetGold,
	// Indices
	"NAS100": AssetIndices, "US30": AssetIndices, "SPX500": AssetIndices,
	"US500": AssetIndices, "DE40": AssetIndices, "UK100": AssetIndices,
	"JP225": AssetIndices, "NDX": AssetIndices, "DJI": AssetIndices,
	// Crypto
	"BTCUSD": AssetCrypto, "ETHUSD": AssetCrypto, "BTCUSDT": AssetCrypto,
	"ETHUSDT": AssetCrypto, "XRPUSD": AssetCrypto, "SOLUSD": AssetCrypto,
	"BNBUSD": AssetCrypto, "DOGEUSD": AssetCrypto,
}

// DetectAssetType guesses the asset type from a symbol string.
// Falls back to Forex if unrecognised.
func DetectAssetType(symbol string) AssetType {
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	if at, ok := knownSymbols[sym]; ok {
		return at
	}
	// Heuristic: 6-char alphabetic → likely forex pair
	if len(sym) == 6 && isAllAlpha(sym) {
		return AssetForex
	}
	return AssetForex
}

func isAllAlpha(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
