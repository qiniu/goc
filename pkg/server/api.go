package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// listServices return all service informations
func (gs *gocServer) listServices(c *gin.Context) {
	services := make([]gocCoveredClient, 0)

	gs.rpcClients.Range(func(key, value interface{}) bool {
		service, ok := value.(gocCoveredClient)
		if !ok {
			return false
		}
		services = append(services, service)
		return true
	})

	c.JSON(http.StatusOK, gin.H{
		"items": services,
	})
}
