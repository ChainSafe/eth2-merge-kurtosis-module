package el_genesis

// Represents the paths to the EL genesis files *on the module container*
type ELGenesisData struct {
	// Path to the directory holding all the EL genesis files
	parentDirpath string
	
	gethGenesisJsonFilepath string

	nethermindGenesisJsonFilepath string

	besuGenesisJsonFilepath string
}

func newELGenesisData(parentDirpath string, gethGenesisJsonFilepath string, nethermindGenesisJsonFilepath string, besuGenesisJsonFilepath string) *ELGenesisData {
	return &ELGenesisData{parentDirpath: parentDirpath, gethGenesisJsonFilepath: gethGenesisJsonFilepath, nethermindGenesisJsonFilepath: nethermindGenesisJsonFilepath, besuGenesisJsonFilepath: besuGenesisJsonFilepath}
}

func (paths *ELGenesisData) GetParentDirpath() string {
	return paths.parentDirpath
}
func (paths *ELGenesisData) GetGethGenesisJsonFilepath() string {
	return paths.gethGenesisJsonFilepath
}
func (paths *ELGenesisData) GetNethermindGenesisJsonFilepath() string {
	return paths.nethermindGenesisJsonFilepath
}
func (paths *ELGenesisData) GetBesuGenesisJsonFilepath() string {
	return paths.besuGenesisJsonFilepath
}