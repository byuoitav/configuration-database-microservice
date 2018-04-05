package handlers

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/configuration-database-microservice/couch"
	"github.com/byuoitav/configuration-database-microservice/log"
	"github.com/labstack/echo"
)

func GetDevicesByRoom(context echo.Context) error {
	roomID := context.Param("roomid")
	buildingID := context.Param("buildingid")
	if len(roomID) < 1 || len(buildingID) < 1 {
		return context.JSON(http.StatusBadRequest, "Need a roomid and buildingid")
	}

	devs, err := couch.GetDevicesByRoom(buildingID + "-" + roomID)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.JSON(http.StatusOK, devs)
}

func GetDeviceByID(context echo.Context) error {
	room := context.Param("roomid")
	building := context.Param("buildingid")
	dev := context.Param("deviceid")
	if len(room) < 1 || len(building) < 1 || len(dev) < 1 {
		return context.JSON(http.StatusBadRequest, "Need a roomid and buildingid")
	}

	device, err := couch.GetDeviceByID(fmt.Sprintf("%v-%v-%v", building, room, dev))
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.JSON(http.StatusOK, device)

}

func CreateDevice(context echo.Context) error {
	return nil
}

func CreateDeviceType(context echo.Context) error {
	return nil
}

func GetDevicesByBuildingAndRoomAndRole(context echo.Context) error {
	room := context.Param("room")
	building := context.Param("building")
	role := context.Param("role")
	if len(room) < 2 || len(building) < 2 || len(role) < 2 {
		msg := fmt.Sprintf("Invalid parameters. Must include a valid buliding, room and role")
		log.L.Warn(msg)
		return context.JSON(http.StatusInternalServerError, msg)
	}

	devs, err := couch.GetDevicesByRoomAndRole(fmt.Sprintf("%v-%v", building, room), role)
	if err != nil {
		msg := fmt.Sprintf("error: %v", err.Error())
		log.L.Warn(msg)
		return context.JSON(http.StatusInternalServerError, err)
	}

	return context.JSON(http.StatusOK, devs)
}
