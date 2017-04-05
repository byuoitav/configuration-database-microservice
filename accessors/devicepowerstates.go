package accessors

import "database/sql"

type DevicePowerState struct {
	ID         int        `json:"id,omitempty"`
	Device     Device     `json:"device"`
	PowerState PowerState `json:"powerstate"`
}

func (accessorGroup *AccessorGroup) GetDevicePowerStates() ([]DevicePowerState, error) {
	rows, err := accessorGroup.Database.Query("SELECT * FROM DevicePowerStates")
	if err != nil {
		return []DevicePowerState{}, err
	}

	devicepowerstates, err := exctractDevicePowerStateData(rows)
	if err != nil {
		return []DevicePowerState{}, err
	}
	defer rows.Close()

	return devicepowerstates, nil
}

func (accessorGroup *AccessorGroup) AddDevicePowerState(dps DevicePowerState) (DevicePowerState, error) {
	response, err := accessorGroup.Database.Exec("INSERT INTO DevicePowerStates (devicePowerStateID, deviceID, powerStateID) VALUES(?,?,?)", dps.ID, dps.Device.ID, dps.PowerState.ID)
	if err != nil {
		return DevicePowerState{}, err
	}

	id, err := response.LastInsertId()
	dps.ID = int(id)

	return dps, nil
}

func exctractDevicePowerStateData(rows *sql.Rows) ([]DevicePowerState, error) {

	var devicepowerstates []DevicePowerState
	var devicepowerstate DevicePowerState
	var id *int
	var dID *int
	var pID *int

	for rows.Next() {
		err := rows.Scan(&id, &dID, &pID)
		if err != nil {
			return []DevicePowerState{}, err
		}

		if id != nil {
			devicepowerstate.ID = *id
		}
		if dID != nil {
			devicepowerstate.Device.ID = *dID
		}
		if pID != nil {
			devicepowerstate.PowerState.ID = *pID
		}

		devicepowerstates = append(devicepowerstates, devicepowerstate)
	}

	err := rows.Err()
	if err != nil {
		return []DevicePowerState{}, err
	}

	return devicepowerstates, nil
}
