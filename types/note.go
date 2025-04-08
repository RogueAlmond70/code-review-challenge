package types

type Note struct {
	ID       string `json:"id"`
	UserId   string `json:"-"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Archived bool   `json:"archived"`
}

type NoteDto struct {
	Title    *string `json:"title"`
	Content  *string `json:"content"`
	Archived *bool   `json:"archived"`
}
