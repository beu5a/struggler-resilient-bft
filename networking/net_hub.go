package networking

import (
	"sync"

	getty "github.com/AlexStocks/getty/transport"
)

type NetworkingHub struct {
	node        *Node                 // Use the fully qualified type name
	connections map[int]*getty.Client // Active connections, keyed by peer ID
	mu          sync.Mutex            // Protects connections
}
