package models

// Report contains the results of a furniture integrity check.
type Report struct {
	TotalExpected      int      `json:"total_expected"`
	TotalFound         int      `json:"total_found"`
	MissingAssets      []string `json:"missing_assets"`
	UnregisteredAssets []string `json:"unregistered_assets"`
	MalformedAssets    []string `json:"malformed_assets"`
	GeneratedAt        string   `json:"generated_at"`
	ExecutionTime      string   `json:"execution_time"`
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
	ID              int    `json:"id"`
	ClassName       string `json:"classname"`
	Revision        int    `json:"revision"`
	Category        string `json:"category"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	AdURL           string `json:"adurl,omitempty"`
	OfferID         int    `json:"offerid,omitempty"`
	Buyout          bool   `json:"buyout,omitempty"`
	RentOfferID     int    `json:"rentofferid,omitempty"`
	RentBuyout      bool   `json:"rentbuyout,omitempty"`
	BC              bool   `json:"bc,omitempty"`
	ExcludedDynamic bool   `json:"excludeddynamic,omitempty"`
	CustomParams    string `json:"customparams,omitempty"`
	SpecialType     int    `json:"specialtype,omitempty"`
	FurniLine       string `json:"furniline,omitempty"`
	Environment     string `json:"environment,omitempty"`
	Rare            bool   `json:"rare,omitempty"`

	// Floor Item Specifics
	DefaultDir int `json:"defaultdir,omitempty"`
	XDim       int `json:"xdim,omitempty"`
	YDim       int `json:"ydim,omitempty"`
	PartColors struct {
		Color []string `json:"color"`
	} `json:"partcolors,omitempty"`
	CanStandOn bool `json:"canstandon,omitempty"`
	CanSitOn   bool `json:"cansiton,omitempty"`
	CanLayOn   bool `json:"canlayon,omitempty"`
}

// Validate checks if the furniture item has the minimum required fields.
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
	return ""
}
