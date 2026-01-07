package models

// FurnitureReport contains the results of a furniture integrity check.
type FurnitureReport struct {
	TotalExpected      int      `json:"total_expected"`
	TotalFound         int      `json:"total_found"`
	MissingAssets      []string `json:"missing_assets"`
	UnregisteredAssets []string `json:"unregistered_assets"`
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

// FurnitureItem represents a single furniture definition
type FurnitureItem struct {
	ID        int    `json:"id"`
	ClassName string `json:"classname"`
}
