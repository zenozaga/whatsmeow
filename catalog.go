package whatsmeow

import (
	"encoding/xml"
	"fmt"

	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

/////////////////////////
/// Catalogs | Methods
/////////////////////////

// Get Catalog list from business account by (JID)
func (cli *Client) GetCatalogs(params types.GetCatalogsParams) (*types.Catalog, error) {

	/// set default limit
	if params.Limit <= 0 {
		params.Limit = 10
	}

	// if params.JID == nil, set client JID as default
	var jid types.JID = cli.getOwnID()

	if params.JID != nil {
		jid = *params.JID
	}

	// if params.Cursor is not nil, set cursor
	var cursor string = ""

	if params.Cursor != nil {
		cursor = *params.Cursor
	}

	// Get the catalogs
	return cli.getCatalogs(jid, params.Limit, cursor)

}

// func (cli *Client) UpdateCatalog(params UpdateCatalogParams) (*waBinary.Node, error) {
// 	return nil, nil
// }

// func (cli *Client) DeleteCatalog(params DeleteCatalogParams) (*waBinary.Node, error) {
// 	return nil, nil
// }

// func (cli *Client) GetCatalog(params GetCatalogParams) (*waBinary.Node, error) {
// 	return nil, nil
// }

////////////////////////
/// Products | Methods
////////////////////////

// / Add product to catalog
func (cli *Client) AddProductToCatalog(params types.CreateProductParams) (*waBinary.Node, error) {
	return nil, nil
}

// /// Update product in catalog
// func (cli *Client) UpdateProductInCatalog(params UpdateProductInCatalogParams) (*waBinary.Node, error) {
// 	return nil, nil
//  }

//  func (cli *Client) DeleteProductInCatalog(params DeleteProductInCatalogParams) (*waBinary.Node, error) {
// 	return nil, nil

//  }

////////////////////////
/// Private Methods ///
////////////////////////

func (cli *Client) getCatalogs(jid types.JID, limit int, cursor string) (*types.Catalog, error) {

	/// Validate the limit
	switch {
	case limit < 10:
		limit = 10
	case limit > 100:
		limit = 100
	}

	queryNodes := []waBinary.Node{

		// Set the limit of items to fetch
		{
			Tag:     "limit",
			Content: limit,
		},

		// Set the width and height of the image
		{
			Tag:     "width",
			Content: 100,
		},

		// Set the width and height of the image
		{
			Tag:     "height",
			Content: 100,
		},
	}

	// if cursor is not empty, add it to the query
	// to fetch the next page
	//
	// cursor is the id of the last item fetched
	if cursor != "" {
		queryNodes = append(queryNodes, waBinary.Node{
			Tag:     "after",
			Content: cursor,
		})
	}

	// Send the request
	result, err := cli.sendIQ(infoQuery{
		Type:      iqGet,
		To:        types.ServerJID,
		Namespace: "w:biz:catalog",
		Content: []waBinary.Node{{
			Tag: "product_catalog",
			Attrs: waBinary.Attrs{
				"jid":               jid,
				"allow_shop_source": true,
			},
			Content: queryNodes,
		}},
	})

	if err != nil {
		return nil, err
	}

	return parseCatalog(result)

}

func parseCatalog(node *waBinary.Node) (*types.Catalog, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}

	catalog := &types.Catalog{}
	product_catalog, ok := node.GetOptionalChildByTag("product_catalog")

	if !ok {
		return nil, fmt.Errorf("Catalog not found")
	}

	err := xml.Unmarshal([]byte(product_catalog.XMLString()), catalog)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("CATALOG", catalog.Product)

	return catalog, nil
}

/*

<iq from="s.whatsapp.net" id="82.224-3" type="result"><product_catalog cart_enabled="true"><source>whatsapp</source><product availability="in stock" is_hidden="false"><status_info><status>APPROVED</status><can_appeal>true</can_appeal></status_info><image_fetch_status>DIRECT_UPLOAD</image_fetch_status><name>Niña</name><description>Prueba</description><id>26948681511397672</id><media><image><original_dimensions height="1600" width="1200"/><request_image_url>https://media-atl3-2.cdn.whatsapp.net/v/t45.5328-4/452287145_977611504116525_8084979750120072954_n.jpg?stp=dst-jpg_p100x100&ccb=1-7&_nc_sid=657aed&_nc_ohc=nx9wE-jAb-oQ7kNvgFc-CB2&_nc_ad=z-m&_nc_cid=0&_nc_ht=media-atl3-2.cdn.whatsapp.net&oh=01_Q5AaIKDtVcFoTOPmyRjIJiQMdd6TOw-SatQqJwlJGUEPy7-o&oe=66A36746</request_image_url><original_image_url>https://media-atl3-2.cdn.whatsapp.net/v/t45.5328-4/452287145_977611504116525_8084979750120072954_n.jpg?ccb=1-7&_nc_sid=657aed&_nc_ohc=nx9wE-jAb-oQ7kNvgFc-CB2&_nc_ad=z-m&_nc_cid=0&_nc_ht=media-atl3-2.cdn.whatsapp.net&oh=01_Q5AaIB_HqNjiZQyPFI1U4JEKCFTpXnoyGlfaXqxFDFDKJ7av&oe=66A36746</original_image_url><id>977611500783192</id></image></media><max_available>99</max_available></product><paging><before></before><after></after></paging></product_catalog></iq>

*/
