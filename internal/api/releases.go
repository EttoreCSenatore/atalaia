package api

import (
	"net/http"

	"github.com/EttoreCSenatore/atalaia/internal/releases"
)

type releasesDTO struct {
	Current  string             `json:"current"`
	Releases []releases.Release `json:"releases"`
}

// releases serves releases.yaml (embedded in the binary) for the "What's new"
// page: every version, newest first, in both languages.
func (s *Server) releases(w http.ResponseWriter, r *http.Request) {
	rs, err := releases.All()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "release notes unavailable")
		return
	}
	writeJSON(w, http.StatusOK, releasesDTO{Current: s.Version, Releases: rs})
}
