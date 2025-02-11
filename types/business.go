package types

import "encoding/xml"

//////////////////////////////
/// Collection | Structs & Types
//////////////////////////////

type BusinessCollectionParams struct {

	// JID is the JID of the WhatsApp business account.
	JID *JID

	Limit     int
	ItemLimit int
}

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
	XMLName     xml.Name  `xml:"product_catalog" json:"-"`
	CartEnabled bool      `xml:"cart_enabled,attr" json:"cart_enabled"`
	Source      string    `xml:"source" json:"source"`
	Paging      Paging    `xml:"paging" json:"paging"`
	Products    []Product `xml:"product" json:"products"`
}

//////////////////////////////
/// Product | Structs & Types
//////////////////////////////

type Product struct {
	XMLName          xml.Name   `xml:"product" json:"-"`
	Availability     string     `xml:"availability,attr" json:"availability"`
	IsHidden         bool       `xml:"is_hidden,attr" json:"is_hidden"`
	StatusInfo       StatusInfo `xml:"status_info" json:"status_info"`
	ImageFetchStatus string     `xml:"image_fetch_status" json:"image_fetch_status"`
	Name             string     `xml:"name" json:"name"`
	Description      string     `xml:"description,omitempty" json:"description,omitempty"`
	ID               string     `xml:"id" json:"id"`
	Media            Media      `xml:"media" json:"media"`
	MaxAvailable     int        `xml:"max_available" json:"max_available"`
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
	Status    string `xml:"status" json:"status"`
	CanAppeal bool   `xml:"can_appeal" json:"can_appeal"`
}

type Media struct {
	Images []Image `xml:"image" json:"images"`
}

type Image struct {
	OriginalDimensions OriginalDimensions `xml:"original_dimensions" json:"original_dimensions"`
	RequestImageURL    string             `xml:"request_image_url" json:"request_image_url"`
	OriginalImageURL   string             `xml:"original_image_url" json:"original_image_url"`
	ID                 string             `xml:"id" json:"id"`
}

type OriginalDimensions struct {
	Height int `xml:"height" json:"height"`
	Width  int `xml:"width" json:"width"`
}

type Paging struct {
	Before string `xml:"before" json:"before"`
	After  string `xml:"after" json:"after"`
}
