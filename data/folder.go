package data

// Folder represents a filesystem folder known to wavepipe
type Folder struct {
	ID       int    `json:"id"`
	ParentID int    `db:"parent_id" json:"parentId"`
	Title    string `json:"title"`
	Path     string `json:"path"`
}

// Subfolders retrieves all folders with this folder as their parent ID
func (f *Folder) Subfolders() ([]Folder, error) {
	return DB.Subfolders(f.ID)
}

// Delete removes an existing Folder from the database
func (f *Folder) Delete() error {
	return DB.DeleteFolder(f)
}

// Load pulls an existing Folder from the database
func (f *Folder) Load() error {
	return DB.LoadFolder(f)
}

// Save creates a new Folder in the database
func (f *Folder) Save() error {
	return DB.SaveFolder(f)
}
