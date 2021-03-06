package accessors

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

//SetDeviceAttribute is used to set a field on the DEVICES TABLE.
func (accessorGroup *AccessorGroup) SetDeviceAttribute(info structs.DeviceAttributeInfo) (structs.Device, error) {

	acceptableColumnNames := make(map[string]string)

	acceptableColumnNames["address"] = "string"
	acceptableColumnNames["input"] = "bool"
	acceptableColumnNames["output"] = "bool"
	acceptableColumnNames["buildingID"] = "int"
	acceptableColumnNames["roomID"] = "int"
	acceptableColumnNames["classID"] = "int"
	acceptableColumnNames["displayName"] = "string"
	acceptableColumnNames["typeID"] = "int"

	if _, ok := acceptableColumnNames[info.AttributeName]; !ok {
		return structs.Device{}, errors.New("invalid column name")
	}

	query := fmt.Sprintf("UPDATE Devices SET %v = ? WHERE deviceID = ?", info.AttributeName)

	var err error
	var res sql.Result

	log.Printf("info: %v", info)

	val := acceptableColumnNames[info.AttributeName]

	if val == "string" {
		fmt.Sprintf("Setting a string value")
		res, err = accessorGroup.Database.Exec(query, info.AttributeValue, info.DeviceID)
		if err != nil {
			return structs.Device{}, err
		}
	} else if val == "int" {
		fmt.Sprintf("Setting a int value")
		var value int
		value, err = strconv.Atoi(info.AttributeValue)
		if err != nil {
			return structs.Device{}, err
		}
		res, err = accessorGroup.Database.Exec(query, value, info.DeviceID)
		if err != nil {
			return structs.Device{}, err
		}
	} else if val == "bool" {
		fmt.Sprintf("Setting a bool value")
		var value bool
		if info.AttributeValue == "false" {
			value = false
		} else if info.AttributeValue == "true" {
			value = true
		} else {
			return structs.Device{}, errors.New("Invalid value for a boolean column")
		}
		res, err = accessorGroup.Database.Exec(query, value, info.DeviceID)
		if err != nil {
			return structs.Device{}, err
		}
	}

	if num, err := res.RowsAffected(); num > 1 || err != nil {
		if err != nil {
			return structs.Device{}, err
		}

		err = errors.New(fmt.Sprintf("There was a problem updating the device type: incorrect number of rows affected: %v. ", num))
		return structs.Device{}, err
	}

	log.Printf("Done.")

	log.Printf("Getting the device to return")
	return accessorGroup.GetDeviceById(info.DeviceID)
}

/*
GetDevicesByQuery is a function that abstracts some of the execution and extraction
of data from the database when we're looking for responses based on the COMPLETE device struct.
The function MAY have the WHERE clause passed in to limit the devices found.
The function MAY have any JOIN clauses necessary to the WEHRE Clause not included in
the base query.
JOIN statements in the base query:
JOIN Rooms on Devices.roomID = Rooms.RoomID
JOIN Buildings on Rooms.buildingID = Buildings.buildingID
JOIN DeviceTypes on Devices.typeID = DeviceTypes.deviceTypeID
If empty string is passed in no WHERE clause will be appended, and thus all devices
will be returned.

Flow	->	Find all devices based on the clause passed in
			->	For each device found find the Ports
			->	For each device found find the Commands

Examples of valid parameters.
Example 1:
`JOIN deviceRole on deviceRole.deviceID = Devices.deviceID
JOIN DeviceRoleDefinition on DeviceRole.deviceRoleDefinitionID = DeviceRoleDefinition.deviceRoleDefinitionID
WHERE DeviceRoleDefinition.name LIKE 'AudioIn'`
Example 2:
`WHERE Devices.RoomID = 1`
*/
func (accessorGroup *AccessorGroup) GetDevicesByQuery(query string, parameters ...interface{}) ([]structs.Device, error) {
	baseQuery := `SELECT DISTINCT Devices.deviceID,
  	Devices.Name as deviceName,
  	Devices.address as deviceAddress,
  	Devices.input,
  	Devices.output,
	Devices.displayName,
  	Rooms.roomID,
  	Rooms.name as roomName,
  	Rooms.description as roomDescription,
	Rooms.roomDesignation as roomDesignation,
  	Buildings.buildingID,
  	Buildings.name as buildingName,
  	Buildings.shortName as buildingShortname,
  	Buildings.description as buildingDescription,
  	DeviceClasses.name as deviceType,
	DeviceTypes.typeName as deviceClass
  	FROM Devices
  	JOIN Rooms on Rooms.roomID = Devices.roomID
  	JOIN Buildings on Buildings.buildingID = Devices.buildingID
  	JOIN DeviceClasses on Devices.classID = DeviceClasses.deviceClassID
	JOIN DeviceTypes on Devices.typeID = DeviceTypes.deviceTypeID
    JOIN DeviceRole on DeviceRole.deviceID = Devices.deviceID
    JOIN DeviceRoleDefinition on DeviceRole.deviceRoleDefinitionID = DeviceRoleDefinition.deviceRoleDefinitionID`

	allDevices := []structs.Device{}

	log.Printf("Making query for devices")
	rows, err := accessorGroup.Database.Query(baseQuery+" "+query, parameters...)
	if err != nil {
		log.Printf("Problem executing query: %v", err.Error())
		return []structs.Device{}, err
	}
	log.Printf("Query executed, evaluating responses")

	defer rows.Close()

	for rows.Next() {

		device := structs.Device{}

		err := rows.Scan(&device.ID,
			&device.Name,
			&device.Address,
			&device.Input,
			&device.Output,
			&device.DisplayName,
			&device.Room.ID,
			&device.Room.Name,
			&device.Room.Description,
			&device.Room.RoomDesignation,
			&device.Building.ID,
			&device.Building.Name,
			&device.Building.Shortname,
			&device.Building.Description,
			&device.Type,
			&device.Class,
		)
		if err != nil {
			return []structs.Device{}, err
		}

		device.Commands, err = accessorGroup.GetDeviceCommandsByBuildingAndRoomAndName(device.Building.Shortname, device.Room.Name, device.Name)
		if err != nil {
			return []structs.Device{}, err
		}

		device.Ports, err = accessorGroup.GetDevicePortsByBuildingAndRoomAndName(device.Building.Shortname, device.Room.Name, device.Name)
		if err != nil {
			return []structs.Device{}, err
		}

		device.PowerStates, err = accessorGroup.GetPowerStatesByDeviceID(device.ID)
		if err != nil {
			return []structs.Device{}, err
		}

		device.Roles, err = accessorGroup.GetRolesByDeviceID(device.ID)
		if err != nil {
			return []structs.Device{}, err
		}

		allDevices = append(allDevices, device)
	}

	return allDevices, nil
}

func (AccessorGroup *AccessorGroup) GetDeviceById(deviceID int) (structs.Device, error) {
	log.Printf("Getting device with deviceID %v", deviceID)

	devices, err := AccessorGroup.GetDevicesByQuery(" WHERE Devices.DeviceID = ?", deviceID)
	if err != nil {
		return structs.Device{}, err
	}
	if len(devices) < 1 {
		return structs.Device{}, errors.New(fmt.Sprintf("No devices found for ID %d", deviceID))
	}

	return devices[0], nil
}

func (AccessorGroup *AccessorGroup) GetDevicesByRoomIdAndRoleId(roomId, roleId int) ([]structs.Device, error) {

	devices, err := AccessorGroup.GetDevicesByQuery("WHERE Rooms.roomID = ? AND DeviceRoleDefinition.deviceRoleDefinitionID = ?", roomId, roleId)
	if err != nil {
		return []structs.Device{}, err
	}

	return devices, nil
}

func (AccessorGroup *AccessorGroup) GetDevicesByRoomId(roomId int) ([]structs.Device, error) {

	devices, err := AccessorGroup.GetDevicesByQuery("WHERE Rooms.roomID = ?", roomId)
	if err != nil {
		return []structs.Device{}, err
	}

	return devices, nil
}

func (AccessorGroup *AccessorGroup) GetRolesByDeviceID(deviceID int) ([]string, error) {
	log.Printf("Getting roles by device ID: %v", deviceID)
	query := `Select DeviceRoleDefinition.name From DeviceRoleDefinition 
	JOIN DeviceRole dr on dr.deviceRoleDefinitionID = DeviceRoleDefinition.deviceRoleDefinitionID 
	WHERE dr.deviceID = ?`

	toReturn := []string{}

	rows, err := AccessorGroup.Database.Query(query, deviceID)
	if err != nil {
		return []string{}, err
	}

	log.Printf("Sheriff, this is no time to panic.")
	defer rows.Close()

	for rows.Next() {
		var value string

		err = rows.Scan(&value)
		if err != nil {
			return []string{}, err
		}

		log.Printf("This is a perfect time to panic")
		log.Printf("value: %s", value)
		toReturn = append(toReturn, value)
	}
	return toReturn, nil
}

//GetPowerStatesByDeviceID gets the powerstates allowed for a given devices based on the
//DevicePowerStates table in the DB.
func (AccessorGroup *AccessorGroup) GetPowerStatesByDeviceID(deviceID int) ([]string, error) {
	query := `SELECT PowerStates.name FROM PowerStates
	JOIN DevicePowerStates on DevicePowerStates.powerStateID = PowerStates.powerStateID
	Where DevicePowerStates.deviceID = ?`

	toReturn := []string{}
	rows, err := AccessorGroup.Database.Query(query, deviceID)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var value string

		err := rows.Scan(&value)
		if err != nil {
			return []string{}, err
		}
		toReturn = append(toReturn, value)
	}
	return toReturn, nil
}

//GetDevicesByBuildingAndRoomAndRole gets the devices in the room specified with the given role,
//as specified in the DeviceRole table in the DB
func (accessorGroup *AccessorGroup) GetDevicesByBuildingAndRoomAndRole(buildingShortname string, roomName string, roleName string) ([]structs.Device, error) {
	log.Printf("Getting ")
	devices, err := accessorGroup.GetDevicesByQuery(`WHERE Rooms.name LIKE ? AND Buildings.shortname LIKE ? AND DeviceRoleDefinition.name LIKE ?`,
		roomName, buildingShortname, roleName)

	if err != nil {
		log.Printf("Error: %v", err.Error())
		return []structs.Device{}, err
	}
	switch strings.ToLower(roleName) {

	}
	return devices, nil

}

//GetDevicesByRoleAndType Gets all teh devices that have a given role and type.
func (accessorGroup *AccessorGroup) GetDevicesByRoleAndType(deviceRole string, deviceType string, production string) ([]structs.Device, error) {
	log.Printf("Making the query")
	return accessorGroup.GetDevicesByQuery(`WHERE DeviceRoleDefinition.name LIKE ? AND DeviceClasses.name LIKE ? AND Rooms.roomDesignation = ?`, deviceRole, deviceType, production)
}

//GetDevicesByBuildingAndRoom get all the devices in the room specified.
func (accessorGroup *AccessorGroup) GetDevicesByBuildingAndRoom(buildingShortname string, roomName string) ([]structs.Device, error) {
	log.Printf("Getting devices in room %s and building %s", roomName, buildingShortname)

	devices, err := accessorGroup.GetDevicesByQuery(
		`WHERE Rooms.name=? AND Buildings.shortName=?`, roomName, buildingShortname)

	if err != nil {
		return []structs.Device{}, err
	}

	return devices, nil
}

//GetDeviceCommandsByBuildingAndRoomAndName gets all the commands for the device
//specified. Note that we assume that device names are unique within a room.
func (accessorGroup *AccessorGroup) GetDeviceCommandsByBuildingAndRoomAndName(buildingShortname string, roomName string, deviceName string) ([]structs.Command, error) {

	log.Printf("Getting all the commands for %v-%v-%v", buildingShortname, roomName, deviceName)
	allCommands := []structs.Command{}
	rows, err := accessorGroup.Database.Query(`SELECT Commands.name as commandName, Endpoints.name as endpointName, Endpoints.path as endpointPath, Microservices.address as microserviceAddress
    FROM Devices
	JOIN DeviceTypes on DeviceTypes.deviceTypeID = Devices.typeID
	JOIN DeviceTypeCommandMapping TypeCommands on TypeCommands.deviceTypeID = DeviceTypes.deviceTypeID
    JOIN Commands on TypeCommands.commandID = Commands.commandID 
	JOIN Endpoints on TypeCommands.endpointID = Endpoints.endpointID 
	JOIN Microservices ON TypeCommands.microserviceID = Microservices.microserviceID
    JOIN Rooms ON Rooms.roomID=Devices.roomID
    JOIN Buildings ON Rooms.buildingID=Buildings.buildingID
    WHERE Rooms.name=? AND Buildings.shortName=? AND Devices.name=?`, roomName, buildingShortname, deviceName)
	if err != nil {
		return []structs.Command{}, err
	}
	defer rows.Close()

	allCommands, err = ExtractCommand(rows)
	if err != nil {
		log.Printf("There was an error with the device commands: %v", err.Error())
	}

	log.Printf("found %v commands", len(allCommands))

	return allCommands, err
}

//GetDevicePortsByBuildingAndRoomAndName gets the ports for the device
//specified. Note that we assume that device names are unique within a room.
/*
 */
func (accessorGroup *AccessorGroup) GetDevicePortsByBuildingAndRoomAndName(buildingShortname string, roomName string, deviceName string) ([]structs.Port, error) {
	allPorts := []structs.Port{}

	rows, err := accessorGroup.Database.Query(`SELECT srcDevice.Name as sourceName, Ports.name as portName, destDevice.Name as DestinationDevice, hostDevice.name as HostDevice FROM Ports
    JOIN PortConfiguration ON Ports.PortID = PortConfiguration.PortID
    JOIN Devices as srcDevice on srcDevice.DeviceID = PortConfiguration.sourceDeviceID
    JOIN Devices as destDevice on destDevice.DeviceID = PortConfiguration.destinationDeviceID
		JOIN Devices as hostDevice on hostDevice.DeviceID = PortConfiguration.hostDeviceID
    JOIN Rooms ON Rooms.roomID=destDevice.roomID
    JOIN Buildings ON Rooms.buildingID=Buildings.buildingID
    WHERE Rooms.name=? AND Buildings.shortName=? AND hostDevice.name=?`, roomName, buildingShortname, deviceName)
	if err != nil {
		log.Print(err)
		return []structs.Port{}, err
	}
	defer rows.Close()

	for rows.Next() {
		port := structs.Port{}

		err := rows.Scan(&port.Source, &port.Name, &port.Destination, &port.Host)
		if err != nil {
			log.Print(err)
			return []structs.Port{}, err
		}

		allPorts = append(allPorts, port)
	}

	return allPorts, nil
}

//GetDeviceByBuildingAndRoomAndName gets the device
//specified. Note that we assume that device names are unique within a room.
func (accessorGroup *AccessorGroup) GetDeviceByBuildingAndRoomAndName(buildingShortname string, roomName string, deviceName string) (structs.Device, error) {
	dev, err := accessorGroup.GetDevicesByQuery("WHERE Buildings.shortName = ? AND Rooms.name = ? AND Devices.name = ?", buildingShortname, roomName, deviceName)
	if err != nil || len(dev) == 0 {
		return structs.Device{}, err
	}

	return dev[0], nil
}

//PutDeviceAttributeByDeviceAndRoomAndBuilding allows you to change attribute values for devices
//Currently sets volume and muted.
func (accessorGroup *AccessorGroup) PutDeviceAttributeByDeviceAndRoomAndBuilding(building string, room string, device string, attribute string, attributeValue string) (structs.Device, error) {
	switch strings.ToLower(attribute) {
	case "volume":
		statement := `update AudioDevices SET volume = ? WHERE deviceID =
			(Select deviceID from Devices
				JOIN Rooms on Rooms.roomID = Devices.roomID
				JOIN Buildings on Buildings.buildingID = Rooms.buildingID
				WHERE Devices.name LIKE ? AND Rooms.name LIKE ? AND Buildings.shortName LIKE ?)`
		val, err := strconv.Atoi(attributeValue)
		if err != nil {
			return structs.Device{}, err
		}

		_, err = accessorGroup.Database.Exec(statement, val, device, room, building)
		if err != nil {
			return structs.Device{}, err
		}
		break

	case "muted":
		var valToSet bool
		switch attributeValue {
		case "true":
			valToSet = true
			break
		case "false":
			valToSet = false
			break
		default:
			return structs.Device{}, errors.New("Invalid attribute value, must be a boolean.")
		}
		statement := `update AudioDevices SET muted = ? WHERE deviceID =
			(Select deviceID from Devices
				JOIN Rooms on Rooms.roomID = Devices.roomID
				JOIN Buildings on Buildings.buildingID = Rooms.buildingID
				WHERE Devices.name LIKE ? AND Rooms.name LIKE ? AND Buildings.shortName LIKE ?)`
		_, err := accessorGroup.Database.Exec(statement, valToSet, device, room, building)
		if err != nil {
			return structs.Device{}, err
		}
		break
	}

	dev, err := accessorGroup.GetDeviceByBuildingAndRoomAndName(building, room, device)
	return dev, err
}

func (accessorGroup *AccessorGroup) AddDevice(d structs.Device) (structs.Device, error) {
	log.Printf("Adding device %v to room %v in building %v", d.Name, d.Room.Name, d.Building.Shortname)

	// get device type string, put it into d.Type
	dt, err := accessorGroup.GetDeviceTypeByName(d.Type)
	if err != nil {
		return structs.Device{}, err
	}

	dc, err := accessorGroup.GetDeviceClassByName(d.Class)
	if err != nil {
		return structs.Device{}, err
	}

	// if device already exists in database, stop
	exists, err := accessorGroup.GetDeviceByBuildingAndRoomAndName(d.Building.Shortname, d.Room.Name, d.Name)
	if err != nil || exists.ID != 0 {
		return structs.Device{}, fmt.Errorf("device already exists in room, please choose a different name")
	}

	// insert into devices
	result, err := accessorGroup.Database.Exec("Insert into Devices (name, address, input, output, buildingID, roomID, classID, typeID, displayName) VALUES (?,?,?,?,?,?,?,?,?)", d.Name, d.Address, d.Input, d.Output, d.Building.ID, d.Room.ID, dt.ID, dc.ID, dc.DisplayName)
	if err != nil {
		return structs.Device{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return structs.Device{}, err
	}

	d.ID = int(id)
	log.Printf(color.HiGreenString("NewID: %v", d.ID))

	// insert the roles into the DeviceRole table
	var deviceroles []structs.DeviceRole
	for _, role := range d.Roles {
		r, err := accessorGroup.GetDeviceRoleDefByName(role)
		if err != nil {
			return structs.Device{}, fmt.Errorf("device role definition: %v does not exist", role)
		}
		var dr structs.DeviceRole
		dr.DeviceID = d.ID
		dr.DeviceRoleDefinitionID = r.ID

		deviceroles = append(deviceroles, dr)
	}

	// insert the powerstates into the DevicePowerStates table
	var devicepowerstates []structs.DevicePowerState
	for _, ps := range d.PowerStates {
		p, err := accessorGroup.GetPowerStateByName(ps)
		if err != nil {
			return structs.Device{}, fmt.Errorf("powerstate: %v does not exist", ps)
		}
		var dps structs.DevicePowerState
		dps.DeviceID = d.ID
		dps.PowerStateID = p.ID

		devicepowerstates = append(devicepowerstates, dps)
	}

	// insert everything else
	for _, dr := range deviceroles {
		_, err = accessorGroup.AddDeviceRole(dr)
		if err != nil {
			return structs.Device{}, err
		}
	}

	for _, ps := range devicepowerstates {
		_, err = accessorGroup.AddDevicePowerState(ps)
		if err != nil {
			return structs.Device{}, err
		}
	}

	// clean up d
	d.Room.Devices = nil
	d.Room.Configuration.Evaluators = nil

	return d, nil
}
