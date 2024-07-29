package types

import "encoding/xml"

//////////////////////////////
/// Catalog | Structs & Types
//////////////////////////////

// GetCatalogsParams is the struct for the parameters of the GetCatalogs function.
type GetCatalogsParams struct {
	// JID is the JID of the WhatsApp business account.
	JID *JID

	Limit  int
	Cursor *string
}

type Catalog struct {
	XMLName     xml.Name `xml:"product_catalog"`
	CartEnabled bool     `xml:"cart_enabled,attr"`
	Source      string   `xml:"source"`
	Product     Product  `xml:"product"`
	Paging      Paging   `xml:"paging"`
}

//////////////////////////////
/// Product | Structs & Types
//////////////////////////////

type Product struct {
	Availability     string     `xml:"availability,attr"`
	IsHidden         bool       `xml:"is_hidden,attr"`
	StatusInfo       StatusInfo `xml:"status_info"`
	ImageFetchStatus string     `xml:"image_fetch_status"`
	Name             string     `xml:"name"`
	Description      string     `xml:"description"`
	ID               string     `xml:"id"`
	Media            Media      `xml:"media"`
	MaxAvailable     int        `xml:"max_available"`
}

type ProductParamsBase struct {
	Name        string
	Description string
	ExternalID  string
	Price       float64
	Currency    string
	Hidden      bool
	Url         *string
}

type CreateProductParams struct {
	Name                string
	Description         string
	ExternalID          string
	Price               float64
	Currency            string
	OriginalCountryCode string
	Hidden              bool
	Images              [][]byte
}

type UpdateProductParams struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Currency    string
	Hidden      bool
}

//////////////////////////////
/// Structs & Types
//////////////////////////////

type StatusInfo struct {
	Status    string `xml:"status"`
	CanAppeal bool   `xml:"can_appeal"`
}

type Media struct {
	Image Image `xml:"image"`
}

type Image struct {
	OriginalDimensions OriginalDimensions `xml:"original_dimensions"`
	RequestImageURL    string             `xml:"request_image_url"`
	OriginalImageURL   string             `xml:"original_image_url"`
	ID                 string             `xml:"id"`
}

type OriginalDimensions struct {
	Height int `xml:"height"`
	Width  int `xml:"width"`
}

type Paging struct {
	Before string `xml:"before"`
	After  string `xml:"after"`
}
