package ole

import (
	"fmt"
	"sync"
	"time"
)

// Connection represents a pooled OLE connection
type Connection struct {
	client     *DbApiClient
	inUse      bool
	lastUsed   time.Time
	id         int
}

// ConnectionPool manages a pool of OLE connections
type ConnectionPool struct {
	connections []*Connection
	mu          sync.Mutex
	maxSize     int
	currentSize int
	companyPath string
}

var (
	globalPool *ConnectionPool
	poolOnce   sync.Once
)

// InitializePool creates the global connection pool
func InitializePool(maxSize int) error {
	var initErr error
	poolOnce.Do(func() {
		if maxSize <= 0 {
			maxSize = 3 // Default to 3 connections
		}
		globalPool = &ConnectionPool{
			connections: make([]*Connection, 0, maxSize),
			maxSize:     maxSize,
			currentSize: 0,
		}
		writeLog(fmt.Sprintf("Connection pool initialized with max size: %d", maxSize))
	})
	return initErr
}

// GetPool returns the global connection pool
func GetPool() *ConnectionPool {
	if globalPool == nil {
		InitializePool(3) // Auto-initialize with default if not done
	}
	return globalPool
}

// Acquire gets an available connection from the pool
func (p *ConnectionPool) Acquire() (*DbApiClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// First, try to find an available existing connection
	for _, conn := range p.connections {
		if !conn.inUse {
			conn.inUse = true
			conn.lastUsed = time.Now()
			writeLog(fmt.Sprintf("Pool: Reusing connection #%d", conn.id))
			
			// Test if connection is still alive
			if conn.client.IsDbcOpen() || p.companyPath == "" {
				return conn.client, nil
			}
			
			// Connection lost, need to reconnect
			writeLog(fmt.Sprintf("Pool: Connection #%d lost, reconnecting...", conn.id))
			if p.companyPath != "" {
				if err := conn.client.OpenDbc(p.companyPath); err != nil {
					writeLog(fmt.Sprintf("Pool: Failed to reconnect #%d: %v", conn.id, err))
					// Remove bad connection
					p.removeConnection(conn)
					// Try to create a new one instead
					return p.createNewConnection()
				}
			}
			return conn.client, nil
		}
	}

	// No available connections, create a new one if under limit
	if p.currentSize < p.maxSize {
		return p.createNewConnection()
	}

	// Pool is full and all connections are in use
	writeLog("Pool: All connections in use, waiting...")
	
	// Wait a bit and try again (simple retry, could be improved with wait queue)
	p.mu.Unlock()
	time.Sleep(100 * time.Millisecond)
	p.mu.Lock()
	
	// Try one more time to find an available connection
	for _, conn := range p.connections {
		if !conn.inUse {
			conn.inUse = true
			conn.lastUsed = time.Now()
			writeLog(fmt.Sprintf("Pool: Reusing connection #%d after wait", conn.id))
			return conn.client, nil
		}
	}
	
	return nil, fmt.Errorf("connection pool exhausted (max: %d)", p.maxSize)
}

// createNewConnection creates a new OLE connection and adds it to the pool
func (p *ConnectionPool) createNewConnection() (*DbApiClient, error) {
	writeLog(fmt.Sprintf("Pool: Creating new connection #%d", p.currentSize+1))
	
	client, err := NewDbApiClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create OLE connection: %w", err)
	}
	
	// Open database if we have a company path
	if p.companyPath != "" {
		if err := client.OpenDbc(p.companyPath); err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
	}
	
	conn := &Connection{
		client:   client,
		inUse:    true,
		lastUsed: time.Now(),
		id:       p.currentSize + 1,
	}
	
	p.connections = append(p.connections, conn)
	p.currentSize++
	
	writeLog(fmt.Sprintf("Pool: Successfully created connection #%d (total: %d)", conn.id, p.currentSize))
	return client, nil
}

// Release returns a connection to the pool
func (p *ConnectionPool) Release(client *DbApiClient) {
	if client == nil {
		return
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for _, conn := range p.connections {
		if conn.client == client {
			conn.inUse = false
			conn.lastUsed = time.Now()
			writeLog(fmt.Sprintf("Pool: Released connection #%d", conn.id))
			return
		}
	}
	
	writeLog("Pool: Warning - released connection not found in pool")
}

// SetCompanyPath sets the company path for all connections
func (p *ConnectionPool) SetCompanyPath(companyPath string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.companyPath = companyPath
	writeLog(fmt.Sprintf("Pool: Company path set to: %s", companyPath))
}

// removeConnection removes a connection from the pool
func (p *ConnectionPool) removeConnection(conn *Connection) {
	for i, c := range p.connections {
		if c == conn {
			// Close the connection
			if c.client != nil {
				c.client.Close()
			}
			// Remove from slice
			p.connections = append(p.connections[:i], p.connections[i+1:]...)
			p.currentSize--
			writeLog(fmt.Sprintf("Pool: Removed connection #%d (remaining: %d)", conn.id, p.currentSize))
			break
		}
	}
}

// CloseAll closes all connections in the pool
func (p *ConnectionPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	writeLog(fmt.Sprintf("Pool: Closing all %d connections", p.currentSize))
	
	for _, conn := range p.connections {
		if conn.client != nil {
			conn.client.Close()
		}
	}
	
	p.connections = nil
	p.currentSize = 0
	writeLog("Pool: All connections closed")
}

// GetStats returns pool statistics
func (p *ConnectionPool) GetStats() (total, inUse, available int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	total = p.currentSize
	for _, conn := range p.connections {
		if conn.inUse {
			inUse++
		}
	}
	available = total - inUse
	return
}

// CleanupIdle removes connections that have been idle for too long
func (p *ConnectionPool) CleanupIdle(maxIdleTime time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	now := time.Now()
	for i := len(p.connections) - 1; i >= 0; i-- {
		conn := p.connections[i]
		if !conn.inUse && now.Sub(conn.lastUsed) > maxIdleTime {
			writeLog(fmt.Sprintf("Pool: Closing idle connection #%d", conn.id))
			conn.client.Close()
			p.connections = append(p.connections[:i], p.connections[i+1:]...)
			p.currentSize--
		}
	}
}