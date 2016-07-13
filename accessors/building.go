package accessors

type Building struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Shortname string `json:"shortname"`
}

// GetAllBuildings returns a list of buildings from the database
func (accessorGroup *AccessorGroup) GetAllBuildings() ([]Building, error) {
	allBuildings := []Building{}

	rows, err := accessorGroup.Database.Query("SELECT * FROM buildings")
	if err != nil {
		return []Building{}, err
	}

	defer rows.Close()

	for rows.Next() {
		building := Building{}

		err := rows.Scan(&building.ID, &building.Name, &building.Shortname)
		if err != nil {
			return []Building{}, err
		}

		allBuildings = append(allBuildings, building)
	}

	err = rows.Err()
	if err != nil {
		return []Building{}, err
	}

	return allBuildings, nil
}

// GetBuildingByID returns a building from the database by ID
func (accessorGroup *AccessorGroup) GetBuildingByID(id int) (Building, error) {
	building := &Building{}
	err := accessorGroup.Database.QueryRow("SELECT * FROM buildings WHERE id=?", id).Scan(&building.ID, &building.Name, &building.Shortname)
	if err != nil {
		return Building{}, err
	}

	return *building, nil
}

// GetBuildingByName returns a building from the database by name
func (accessorGroup *AccessorGroup) GetBuildingByName(name string) (Building, error) {
	building := &Building{}
	err := accessorGroup.Database.QueryRow("SELECT * FROM buildings WHERE name=?", name).Scan(&building.ID, &building.Name, &building.Shortname)
	if err != nil {
		return Building{}, err
	}

	return *building, nil
}

// GetBuildingByShortname returns a building from the database by shortname
func (accessorGroup *AccessorGroup) GetBuildingByShortname(shortname string) (Building, error) {
	building := &Building{}
	err := accessorGroup.Database.QueryRow("SELECT * FROM buildings WHERE shortname=?", shortname).Scan(&building.ID, &building.Name, &building.Shortname)
	if err != nil {
		return Building{}, err
	}

	return *building, nil
}

// MakeBuilding adds a building to the database
func (accessorGroup *AccessorGroup) MakeBuilding(name string, shortname string) (Building, error) {
	_, err := accessorGroup.Database.Query("INSERT INTO buildings (name, shortname) VALUES (?, ?)", name, shortname)
	if err != nil {
		return Building{}, err
	}

	building, err := accessorGroup.GetBuildingByName(name)
	if err != nil {
		return Building{}, err
	}

	return building, nil
}
