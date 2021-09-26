package server

type ReindexReq struct {
	Source Index `json:"source"`
	Dest   Index `json:"dest"`
}

type Index struct {
	Index string `json:"index"`
}

type OK struct {
	ID string `json:"id"`
}

type Err struct {
	Msg string `json:"msg"`
}
