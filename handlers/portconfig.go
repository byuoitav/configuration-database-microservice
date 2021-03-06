package handlers

import (
	"net/http"
	"strconv"

	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/labstack/echo"
)

func (handlerGroup *HandlerGroup) AddPortConfiguration(context echo.Context) error {
	pcID := context.Param("id")
	var pc structs.PortConfiguration

	err := context.Bind(&pc)
	if err != nil {
		return context.JSON(http.StatusBadRequest, err.Error())
	}
	// have to convert pc.ID to a string to compare it to a string (pcID)
	if pcID != strconv.Itoa(pc.ID) {
		return context.JSON(http.StatusBadRequest, "Endpoint parameter and json id must match!")
	}

	response, err := handlerGroup.Accessors.AddPortConfiguration(pc)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.JSON(http.StatusOK, response)
}
