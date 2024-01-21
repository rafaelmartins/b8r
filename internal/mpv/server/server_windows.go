package server

func getSocket(id string) string {
	if id == "" {
		id = "UNK"
	}
	return `\\.\pipe\b8r-mpv-` + id
}
