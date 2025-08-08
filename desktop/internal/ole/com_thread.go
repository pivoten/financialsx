package ole

import (
	"fmt"
	"runtime"
	"sync"
)

// comRequest represents a request to execute on the COM thread
type comRequest struct {
	companyPath string
	fn          func(*DbApiClient) error
	result      error
	done        chan bool
}

var (
	comClient   *DbApiClient
	comRequests = make(chan *comRequest, 100)
	comStarted  sync.Once
	comShutdown = make(chan bool)
)

// initCOMThread starts the dedicated COM thread
func initCOMThread() {
	comStarted.Do(func() {
		go runCOMThread()
	})
}

// runCOMThread runs on a dedicated OS thread for COM operations
func runCOMThread() {
	// Lock this goroutine to its OS thread
	// This is CRITICAL for COM - all COM calls must happen on the same thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	
	writeLog("COM Thread: Starting dedicated COM thread")
	
	// Create the OLE client on this thread
	var err error
	comClient, err = NewDbApiClient()
	if err != nil {
		writeLog(fmt.Sprintf("COM Thread: Failed to create OLE client: %v", err))
		// Handle all pending requests with error
		for {
			select {
			case req := <-comRequests:
				req.result = fmt.Errorf("COM initialization failed: %w", err)
				req.done <- true
			case <-comShutdown:
				return
			default:
				return
			}
		}
	}
	
	writeLog("COM Thread: OLE client created successfully")
	currentPath := ""
	
	// Process requests
	for {
		select {
		case req := <-comRequests:
			writeLog(fmt.Sprintf("COM Thread: Processing request for company: %s", req.companyPath))
			
			// Switch database if needed
			if req.companyPath != "" && req.companyPath != currentPath {
				writeLog(fmt.Sprintf("COM Thread: Switching database from '%s' to '%s'", currentPath, req.companyPath))
				
				// Close current database if open
				if currentPath != "" {
					comClient.CloseDbc()
				}
				
				// Open new database
				err := comClient.OpenDbc(req.companyPath)
				if err != nil {
					writeLog(fmt.Sprintf("COM Thread: Failed to open database: %v", err))
					req.result = fmt.Errorf("failed to open database: %w", err)
					req.done <- true
					continue
				}
				currentPath = req.companyPath
			}
			
			// Execute the request
			req.result = req.fn(comClient)
			req.done <- true
			
		case <-comShutdown:
			writeLog("COM Thread: Shutting down")
			if comClient != nil {
				comClient.Close()
				comClient = nil
			}
			return
		}
	}
}

// ExecuteOnCOMThread executes a function on the dedicated COM thread
func ExecuteOnCOMThread(companyPath string, fn func(*DbApiClient) error) error {
	// Ensure COM thread is started
	initCOMThread()
	
	// Create request
	req := &comRequest{
		companyPath: companyPath,
		fn:          fn,
		done:        make(chan bool, 1),
	}
	
	// Send request
	comRequests <- req
	
	// Wait for completion
	<-req.done
	
	return req.result
}

// ShutdownCOMThread shuts down the COM thread
func ShutdownCOMThread() {
	writeLog("Shutting down COM thread")
	close(comShutdown)
}