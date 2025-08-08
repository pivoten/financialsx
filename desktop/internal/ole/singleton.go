package ole

import (
	"fmt"
	"sync"
)

var (
	singletonClient *DbApiClient
	singletonMutex  sync.Mutex
	singletonOnce   sync.Once
)

// GetSingletonClient returns the single global OLE connection
func GetSingletonClient() (*DbApiClient, error) {
	var err error
	
	singletonOnce.Do(func() {
		writeLog("Singleton: Creating single OLE connection")
		singletonClient, err = NewDbApiClient()
		if err != nil {
			writeLog(fmt.Sprintf("Singleton: Failed to create connection: %v", err))
		} else {
			writeLog("Singleton: Connection created successfully")
		}
	})
	
	if err != nil {
		return nil, err
	}
	
	if singletonClient == nil {
		return nil, fmt.Errorf("singleton connection is nil")
	}
	
	return singletonClient, nil
}

// WithSingletonClient executes a function with the singleton client
// This ensures thread-safe access to the single connection
func WithSingletonClient(companyPath string, fn func(*DbApiClient) error) error {
	singletonMutex.Lock()
	defer singletonMutex.Unlock()
	
	client, err := GetSingletonClient()
	if err != nil {
		return fmt.Errorf("failed to get singleton client: %w", err)
	}
	
	// Ensure the correct database is open
	if companyPath != "" {
		currentPath := client.GetDbcPath()
		if currentPath != companyPath {
			writeLog(fmt.Sprintf("Singleton: Switching database from %s to %s", currentPath, companyPath))
			client.CloseDbc()
			if err := client.OpenDbc(companyPath); err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
		}
	}
	
	// Execute the function with the client
	return fn(client)
}

// CloseSingleton closes the singleton connection
func CloseSingleton() {
	singletonMutex.Lock()
	defer singletonMutex.Unlock()
	
	if singletonClient != nil {
		writeLog("Singleton: Closing connection")
		singletonClient.Close()
		singletonClient = nil
	}
}