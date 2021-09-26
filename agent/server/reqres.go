package server

type ReindexReq struct {
	Source Index `json:"source"`
	Dest   Index `json:"dest"`
}

type Index struct {
	Index string `json:"index"`
}
type HealthzOK struct {
	Msg string `json:"msg"`
}

type ReindexOK struct {
	ID string `json:"id"`
}

type ReindexErr struct {
	Msg string `json:"msg"`
}
