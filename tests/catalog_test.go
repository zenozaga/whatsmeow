package tests

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/lib/pq"
)

func newPostgresDBInstance(database_url string, dbLog waLog.Logger) *sqlstore.Container {

	uri, err := url.Parse(database_url)

	if err != nil {
		panic(err)
	}

	/// get port from url
	port, err := strconv.Atoi(uri.Port())

	if err != nil {
		panic(err)
	}

	username := uri.User.Username()
	password, _ := uri.User.Password()
	schema := uri.Query().Get("schema")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		uri.Hostname(), port, username, password, uri.Path[1:], schema)

	// "postgresql://postgres:0121@localhost:5432/automatizadovip?schema=whatsapp"
	// psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	// 	"localhost", 5432, "postgres", "", 0121"automatizadovip")

	container, err := sqlstore.New("postgres", psqlInfo, dbLog)
	if err != nil {
		panic(err)
	}
	return container

}

func TestCatalogs(t *testing.T) {

	externalID := "madelyn"

	/// logger
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container := newPostgresDBInstance("postgresql://postgres:0121@10.0.0.195:5432/automatizadovip?schema=whastapp", dbLog)
	device, err := container.GetDeviceByExternalID(externalID)

	if err != nil {
		panic(err)
	}

	client := whatsmeow.NewClient(device, dbLog)

	t.Run("Connection", func(t *testing.T) {

		// try to connect
		connected := whatsmeow.TryConnectClient(client, 5*time.Second)

		if !connected {
			t.Errorf("Client could not connect")
			return
		}

		t.Log("Client connected")

	})

	t.Run("GetCatalogs", func(t *testing.T) {

		catalog, err := client.GetCatalogs(types.GetCatalogsParams{})
		if err != nil {
			t.Errorf("Error fetching catalogs: %v", err)
			t.FailNow()
		}

		fmt.Println("Fetched Successfully", catalog)

		t.Log("Fetched Successfully")

	})

}
