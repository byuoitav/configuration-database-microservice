package accessors

import (
	"database/sql"
	"errors"
	"log"
)

//Room represents a room object as represented in the DB.
type Room struct {
	ID                 int               `json:"id,omitempty"`
	Name               string            `json:"name,omitempty"`
	Description        string            `json:"description,omitempty"`
	Building           Building          `json:"building,omitempty"`
	CurrentVideoInput  int               `json:"currentVideoInput,omitempty"`
	CurrentAudioInput  int               `json:"currentAudioInput,omitempty"`
	CurrentVideoOutput int               `json:"currentVideoOutput,omitempty"`
	CurrentAudioOutput int               `json:"currentAudioOutput,omitempty"`
	Devices            []Device          `json:"devices,omitempty"`
	ConfigurationID    int               `json:"configurationID,omitempty"`
	Configuration      RoomConfiguration `json:"configuration"`
}

// GetAllRooms returns a list of rooms from the database
func (accessorGroup *AccessorGroup) GetAllRooms() ([]Room, error) {
	allBuildings := []Building{}

	rows, err := accessorGroup.Database.Query("SELECT * FROM Buildings")
	if err != nil {
		return []Room{}, err
	}

	for rows.Next() {
		building := Building{}

		err = rows.Scan(&building.ID, &building.Name, &building.Shortname)
		if err != nil {
			return []Room{}, err
		}

		allBuildings = append(allBuildings, building)
	}

	allRooms := []Room{}

	rows, err = accessorGroup.Database.Query("SELECT * FROM Rooms")
	if err != nil {
		return []Room{}, err
	}

	defer rows.Close()

	for rows.Next() {
		room := Room{}

		err = rows.Scan(&room.ID, &room.Name, &room.Building.ID, &room.Description)
		if err != nil {
			return []Room{}, err
		}

		for i := 0; i < len(allBuildings); i++ {
			if allBuildings[i].ID == room.Building.ID {
				room.Building = allBuildings[i]
				break
			}
		}

		allRooms = append(allRooms, room)
	}

	err = rows.Err()
	if err != nil {
		return []Room{}, err
	}

	return allRooms, nil
}

// GetRoomByID returns a room from the database by ID
func (accessorGroup *AccessorGroup) GetRoomByID(id int) (Room, error) {
	room := &Room{}

	err := accessorGroup.Database.QueryRow("SELECT * FROM rooms WHERE id=?", id).Scan(&room.ID, &room.Name, &room.Building.ID, &room.Description)
	if err != nil {
		return Room{}, err
	}

	return *room, nil
}

//ExtractRoomData pulls data from a sql query
func (accessorGroup *AccessorGroup) ExtractRoomData(rows *sql.Rows) (rooms []Room, err error) {

	for rows.Next() {
		room := Room{}

		err = rows.Scan(
			&room.ID,
			&room.Name,
			&room.Building.ID,
			&room.Description,
			&room.CurrentVideoOutput,
			&room.CurrentAudioOutput,
			&room.CurrentVideoInput,
			&room.CurrentAudioInput,
			&room.ConfigurationID,
		)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			return
		}
		rooms = append(rooms, room)
	}
	return
}

// GetRoomsByBuilding returns a room from the database by building
func (accessorGroup *AccessorGroup) GetRoomsByBuilding(building string) ([]Room, error) {

	rows, err := accessorGroup.Database.Query(`SELECT * FROM Rooms
		JOIN Buildings ON Rooms.buildingID = Buildings.buildingID WHERE buildingShortname=?`, building)
	if err != nil {
		return []Room{}, err
	}
	allRooms, err := accessorGroup.ExtractRoomData(rows)
	if err != nil {
		return []Room{}, err
	}
	return allRooms, nil
}

// GetRoomByBuildingAndName returns a room from the database by building shortname and room name
func (accessorGroup *AccessorGroup) GetRoomByBuildingAndName(buildingShortname string, name string) (Room, error) {
	log.Printf("Getting room info for %s - %s...", buildingShortname, name)
	building, err := accessorGroup.GetBuildingByShortname(buildingShortname)
	if err != nil {
		return Room{}, err
	}

	room := Room{}
	room.Building = building
	log.Printf("Getting room info for %s-%s...", buildingShortname, name)
	row, err := accessorGroup.Database.Query("SELECT * FROM Rooms WHERE buildingID=? AND name=?", building.ID, name)
	if err != nil {
		return Room{}, err
	}
	rooms, err := accessorGroup.ExtractRoomData(row)
	if err != nil {
		return Room{}, err
	}
	room = rooms[0]

	log.Printf("Getting device info for %s-%s...", buildingShortname, name)
	room.Devices, err = accessorGroup.GetDevicesByBuildingAndRoom(buildingShortname, name)
	if err != nil {
		return room, err
	}

	log.Printf("Gettign configuration information for %s-%s...", buildingShortname, name)
	room.Configuration, err = accessorGroup.GetConfigurationByConfigurationID(room.ConfigurationID)
	if err != nil {
		return room, err
	}

	log.Printf("Done.")
	return room, nil
}

// MakeRoom adds a room to the database
func (accessorGroup *AccessorGroup) MakeRoom(name string, buildingShortname string, vlan int) (Room, error) {
	building, err := accessorGroup.GetBuildingByShortname(buildingShortname)
	if err != nil {
		return Room{}, errors.New("Could not find a building with the \"" + buildingShortname + "\" shortname")
	}

	_, err = accessorGroup.Database.Exec("INSERT INTO rooms (name, building, vlan) VALUES (?, ?, ?)", name, building.ID, vlan)
	if err != nil {
		return Room{}, err
	}

	room, err := accessorGroup.GetRoomByBuildingAndName(building.Shortname, name)
	if err != nil {
		return Room{}, err
	}

	return room, nil
}
