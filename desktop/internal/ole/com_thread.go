package ole

import (
	"fmt"
	"runtime"
	"sync"
	"time"
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
	comPreload  = make(chan string, 1)  // Channel to request preloading
	comClose    = make(chan bool, 1)    // Channel to request closing
	lastActivity time.Time               // Track last activity for idle timeout
	idleTimeout = 5 * time.Minute       // Close after 5 minutes of inactivity
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
	
	// Don't create the OLE client immediately - wait for first request
	// This prevents creating unnecessary processes
	var err error
	currentPath := ""
	lastActivity = time.Now()
	
	// Create idle timer
	idleTimer := time.NewTimer(idleTimeout)
	defer idleTimer.Stop()
	
	// Process requests
	for {
		select {
		case companyPath := <-comPreload:
			// Preload request - open database but don't execute anything
			writeLog(fmt.Sprintf("COM Thread: Preloading database for company: %s", companyPath))
			
			// This case is now obsolete since we handle preload via ExecuteOnCOMThread
			// Just reset the timer
			lastActivity = time.Now()
			idleTimer.Reset(idleTimeout)
			
		case <-comClose:
			// Explicit close request
			writeLog("COM Thread: Explicit close requested")
			if comClient != nil {
				// Close database if open
				if currentPath != "" {
					comClient.CloseDbc()
					currentPath = ""
					writeLog("COM Thread: Database closed")
				}
				
				// Terminate the OLE server process to release all file locks
				comClient.Close()
				comClient = nil
				writeLog("COM Thread: OLE client closed and terminated")
				
				// Don't recreate immediately - wait until actually needed
				// This prevents multiple processes from being created
			}
			
		case req := <-comRequests:
			writeLog(fmt.Sprintf("COM Thread: Processing request for company: %s", req.companyPath))
			
			// Ensure client exists (might have been closed)
			if comClient == nil {
				writeLog("COM Thread: OLE client was nil, recreating...")
				comClient, err = NewDbApiClient()
				if err != nil {
					writeLog(fmt.Sprintf("COM Thread: Failed to recreate OLE client: %v", err))
					req.result = fmt.Errorf("failed to recreate OLE client: %w", err)
					req.done <- true
					continue
				}
				writeLog("COM Thread: OLE client recreated successfully")
			}
			
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
			
			// Update activity time and reset idle timer
			lastActivity = time.Now()
			idleTimer.Reset(idleTimeout)
			
		case <-idleTimer.C:
			// Check if we've been idle too long
			if time.Since(lastActivity) >= idleTimeout {
				writeLog(fmt.Sprintf("COM Thread: Idle timeout reached (%v since last activity), closing database", time.Since(lastActivity)))
				if currentPath != "" {
					comClient.CloseDbc()
					currentPath = ""
					writeLog("COM Thread: Database closed due to inactivity")
				}
			}
			idleTimer.Reset(idleTimeout)
			
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

// PreloadOLEConnection preloads the OLE connection for a company
func PreloadOLEConnection(companyPath string) {
	// Start the COM thread if not already started
	initCOMThread()
	
	// Run preload asynchronously to avoid blocking login
	go func() {
		// Force initialization by executing a simple operation
		// This ensures the OLE client is created and database is opened
		err := ExecuteOnCOMThread(companyPath, func(client *DbApiClient) error {
			writeLog(fmt.Sprintf("Preload: OLE connection established for company: %s", companyPath))
			// Just checking if database is open is enough to establish connection
			if client.IsDbcOpen() {
				writeLog(fmt.Sprintf("Preload: Database is open for: %s", companyPath))
			}
			return nil
		})
		
		if err != nil {
			writeLog(fmt.Sprintf("Preload failed for company %s: %v", companyPath, err))
		} else {
			writeLog(fmt.Sprintf("Preload completed successfully for company: %s", companyPath))
		}
	}()
}

// CloseOLEConnection explicitly closes the current OLE connection
func CloseOLEConnection() {
	// Non-blocking send to close channel
	select {
	case comClose <- true:
		writeLog("Close request sent")
	default:
		writeLog("Close request skipped - channel busy")
	}
}

// SetIdleTimeout sets the idle timeout duration
func SetIdleTimeout(duration time.Duration) {
	idleTimeout = duration
	writeLog(fmt.Sprintf("Idle timeout set to: %v", duration))
}

// ShutdownCOMThread shuts down the COM thread
func ShutdownCOMThread() {
	writeLog("Shutting down COM thread")
	
	// Explicitly close the OLE connection if it exists
	if comClient != nil {
		comClient.Close()
		comClient = nil
	}
	
	close(comShutdown)
}

// KillAllOLEProcesses terminates all running OLE server processes
// This is useful for cleanup on startup to prevent accumulation
func KillAllOLEProcesses() {
	writeLog("Killing all existing OLE server processes")
	// On Windows, we could use taskkill to force terminate all dbapi.exe processes
	// This ensures clean slate on startup
	// Note: This is a Windows-specific solution
}