package whatsmeow

import (
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

// TryConnectClient attempts to connect the client to the server asynchronously.
// This is useful for connecting to the server with a saved client.
func TryConnectClient(cli *Client, timeout time.Duration) bool {
	// If ID is not stored, the client was never connected.
	if cli.Store.ID == nil {
		return false
	}

	// Check if the client is already connected before attempting to connect.
	if cli.IsConnected() {
		return true
	}

	result := make(chan bool)

	// Attempt to connect the client asynchronously.
	go func(conn *Client, channel chan bool) {
		// Event handler to handle connection events.
		eventHandler := func(evt interface{}) {
			switch evt.(type) {
			case *events.Connected:
				// Connection successful.
				channel <- true
			case *events.LoggedOut, *events.ConnectFailure:
				// Connection failed or closed.
				channel <- false
			}
		}

		// Add event handler.
		handlerID := conn.AddEventHandler(eventHandler)

		// Attempt to connect.
		err := conn.Connect()
		if err != nil {
			// Connection attempt failed.
			channel <- false
			conn.RemoveEventHandler(handlerID)
			return
		}

		// Wait for the connection result or timeout.
		select {
		case res := <-channel:
			// Connection result received.
			conn.RemoveEventHandler(handlerID)
			channel <- res
		case <-time.After(timeout):
			// Connection attempt timed out.
			conn.RemoveEventHandler(handlerID)
			channel <- false
		}
	}(cli, result)

	// Return the connection result.
	return <-result
}
