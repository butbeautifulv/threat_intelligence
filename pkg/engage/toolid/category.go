package toolid

// Category matches HexStrike tool groupings.
type Category string

const (
	CategoryNetwork Category = "network"
	CategoryWeb     Category = "web"
	CategoryCloud   Category = "cloud"
	CategoryBinary  Category = "binary"
	CategoryAuth    Category = "auth"
	CategoryOSINT   Category = "osint"
	CategoryCTF     Category = "ctf"
	CategoryIntel   Category = "intelligence"
)

func (c Category) Valid() bool {
	switch c {
	case CategoryNetwork, CategoryWeb, CategoryCloud, CategoryBinary,
		CategoryAuth, CategoryOSINT, CategoryCTF, CategoryIntel:
		return true
	default:
		return false
	}
}
