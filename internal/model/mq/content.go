package mq

type MetaContentCdcJson struct {
	CanalJson
	Data []MetaContent `json:"data"`
}

type PublicContentCdcJson struct {
	CanalJson
	Data []PublicContent `json:"data"`
}

type MetaContent struct {
	Id              string `json:"id"`
	Version         string `json:"version"`
	Status          string `json:"status"`
	OldStatus       string `json:"old_status"`
	Desc            string `json:"desc"`
	UserId          string `json:"userid"`
	Title           string `json:"title"`
	PhotoUriList    string `json:"photo_uri_list"`
	ShortText       string `json:"short_text"`
	LongTextUri     string `json:"long_text_uri"`
	VideoUriList    string `json:"video_uri_list"`
	OldPhotoUriList string `json:"old_photo_uri_list"`
	OldShortTextUri string `json:"old_short_text_uri"`
	OldLongTextUri  string `json:"old_long_text_uri"`
	OldVideoUriList string `json:"old_video_uri_list"`
}

type PublicContent struct {
	Id           string `json:"id"`
	Version      string `json:"version"`
	UserId       string `json:"userid"`
	Status       string `json:"status"`
	Title        string `json:"title"`
	PhotoUriList string `json:"photo_uri_list"`
	ShortTextUri string `json:"short_text_uri"`
	LongTextUri  string `json:"long_text_uri"`
	VideoUriList string `json:"video_uri_list"`
	UpdatedAt    string `json:"updated_at"`
	CreatedAt    string `json:"created_at"`
}
