package types

type Note struct {
	ID       string `json:"id"`
	UserId   string `json:"-"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Archived bool   `json:"archived"`
}

type NoteDto struct { // Note Data Transfer Object is a generic and useless name. Need to clarify exactly what it's for, and name it accordingly.
	Title    *string `json:"title"`
	Content  *string `json:"content"`
	Archived *bool   `json:"archived"`
}

type NotesResponse struct {
	Notes      []Note `json:"notes"`
	TotalNotes int    `json:"total notes"`
	Offset     int    `json:"offset"`
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"hasMore"`
}
