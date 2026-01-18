package models

import (
	"strings"
)

// Report contains the results of a furniture integrity check.
type Report struct {
	Assets           []AssetIntegrityItem `json:"assets"`
	TotalAssets      int                  `json:"total_assets"`
	StorageMissing   int                  `json:"storage_missing"`
	DatabaseMissing  int                  `json:"database_missing"`
	FurniDataMissing int                  `json:"furnidata_missing"`
	WithMismatches   int                  `json:"with_mismatches"`
	GeneratedAt      string               `json:"generated_at"`
	ExecutionTime    string               `json:"execution_time"`
}

// AssetIntegrityItem represents the integrity status of a single furniture asset.
type AssetIntegrityItem struct {
	ID               int      `json:"id,omitempty"`
	Name             string   `json:"name"`
	ClassName        string   `json:"class_name,omitempty"`
	FurniDataMissing bool     `json:"furnidata_missing"`
	StorageMissing   bool     `json:"storage_missing"`
	DatabaseMissing  bool     `json:"database_missing"`
	Mismatches       []string `json:"mismatches,omitempty"`
}

// FurnitureDetailReport contains the detailed integrity check for a single item.
type FurnitureDetailReport struct {
	ID              int      `json:"id"`
	ClassName       string   `json:"class_name"`
	Name            string   `json:"name"`
	NitroFile       string   `json:"nitro_file,omitempty"`
	FileExists      bool     `json:"file_exists"`
	InFurniData     bool     `json:"in_furnidata"`
	InDB            bool     `json:"in_db"`
	IntegrityStatus string   `json:"integrity_status"` // "PASS", "FAIL", "WARNING"
	Mismatches      []string `json:"mismatches,omitempty"`
}

// FurnitureData represents the structure of FurniData.json
type FurnitureData struct {
	RoomItemTypes struct {
		FurniType []FurnitureItem `json:"furnitype"`
	} `json:"roomitemtypes"`
	WallItemTypes struct {
		FurniType []FurnitureItem `json:"furnitype"`
	} `json:"wallitemtypes"`
}

// FurnitureItem represents a single furniture definition matching FURNIDATA.md
type FurnitureItem struct {
	// Common Parameters
	ID              int    `json:"id" gamedata:"id"`
	ClassName       string `json:"classname" gamedata:"classname"`
	Revision        int    `json:"revision" gamedata:"revision"`
	Category        string `json:"category" gamedata:"category"`
	Name            string `json:"name" gamedata:"name"`
	Description     string `json:"description" gamedata:"description"`
	AdURL           string `json:"adurl,omitempty" gamedata:"adurl"`
	OfferID         int    `json:"offerid,omitempty" gamedata:"offerid"`
	Buyout          bool   `json:"buyout,omitempty" gamedata:"buyout"`
	RentOfferID     int    `json:"rentofferid,omitempty" gamedata:"rentofferid"`
	RentBuyout      bool   `json:"rentbuyout,omitempty" gamedata:"rentbuyout"`
	BC              bool   `json:"bc,omitempty" gamedata:"bc"`
	ExcludedDynamic bool   `json:"excludeddynamic,omitempty" gamedata:"excludeddynamic"`
	CustomParams    string `json:"customparams,omitempty" gamedata:"customparams"`
	SpecialType     int    `json:"specialtype,omitempty" gamedata:"specialtype"`
	FurniLine       string `json:"furniline,omitempty" gamedata:"furniline"`
	Environment     string `json:"environment,omitempty" gamedata:"environment"`
	Rare            bool   `json:"rare,omitempty" gamedata:"rare"`

	// Floor Item Specifics
	DefaultDir int `json:"defaultdir,omitempty" gamedata:"defaultdir"`
	XDim       int `json:"xdim,omitempty" gamedata:"xdim"`
	YDim       int `json:"ydim,omitempty" gamedata:"ydim"`
	PartColors struct {
		Color []string `json:"color"`
	} `json:"partcolors,omitempty" gamedata:"partcolors"`
	CanStandOn bool `json:"canstandon,omitempty" gamedata:"canstandon"`
	CanSitOn   bool `json:"cansiton,omitempty" gamedata:"cansiton"`
	CanLayOn   bool `json:"canlayon,omitempty" gamedata:"canlayon"`
}

// Validate checks if the furniture item has the minimum required fields and valid formats.
func (i FurnitureItem) Validate() string {
	if i.ID == 0 {
		return "missing id"
	}
	if i.ClassName == "" {
		return "missing classname"
	}
	if i.Name == "" {
		return "missing name"
	}
	if i.Revision == 0 {
		// Revision 0 is technically possible but usually it starts at 1?
		// Docs say "Asset version number". If missing, what happens?
		// Let's assume it should be present. But maybe 0 is valid.
		// However, missing in JSON implies 0 in int.
		// Let's warn if 0? Or just "missing revision" if we treat 0 as default/missing.
		// Many ancient assets might not have it tailored?
		// Safest is to not enforce Revision != 0 unless strictly known.
		// But let's enforce provided fields.
	}
	if i.Category == "" {
		// Category seems important for classification.
		// "General classification tag".
		return "missing category"
	}

	// Validate ClassName format
	// Format: base_name or base_name*color_id
	if strings.Contains(i.ClassName, "*") {
		parts := strings.Split(i.ClassName, "*")
		if len(parts) != 2 {
			return "invalid classname format: too many asterisks"
		}
		if parts[0] == "" {
			return "invalid classname format: empty base name"
		}
		if parts[1] == "" {
			return "invalid classname format: empty color index"
		}
		// Check if color index is numeric?
		// "parses the suffix as the colorIndex (variable)"
		// Typically numeric, but effectively a string in the name?
		// Docs say "color_id". Example "chair_wood*1".
		// Let's not be too strict on the ID content unless we know it must be int.
	}

	return ""
}
