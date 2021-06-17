package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/goc/v2/pkg/log"
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

// getProfiles get and merge all services' informations
func (gs *gocServer) getProfiles(c *gin.Context) {
	gs.rpcClients.Range(func(key, value interface{}) bool {
		service, ok := value.(gocCoveredClient)
		if !ok {
			return false
		}

		var req GetProfileReq = "getprofile"
		var res GetProfileRes
		err := service.rpc.Call("GocAgent.GetProfile", req, &res)
		if err != nil {
			log.Errorf("fail to get profile from: %v, reasson: %v", service.Id, err)
			return true
		}
		log.Infof("res: %v", res)

		return true
	})
}
