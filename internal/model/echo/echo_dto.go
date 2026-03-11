package model

import fileModel "github.com/lin-snow/ech0/internal/model/file"

type EchoExtensionDto struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type EchoUpsertDto struct {
	ID        string               `json:"id,omitempty"`
	Content   string               `json:"content"`
	EchoFiles []fileModel.EchoFile `json:"echo_files,omitempty"`
	Layout    string               `json:"layout,omitempty"`
	Private   bool                 `json:"private"`
	Extension *EchoExtensionDto    `json:"extension,omitempty"`
	Tags      []Tag                `json:"tags,omitempty"`
}

func (dto *EchoUpsertDto) ToModel() *Echo {
	echo := &Echo{
		ID:        dto.ID,
		Content:   dto.Content,
		EchoFiles: dto.EchoFiles,
		Layout:    dto.Layout,
		Private:   dto.Private,
		Tags:      dto.Tags,
	}
	if dto.Extension != nil {
		echo.Extension = &EchoExtension{
			Type:    dto.Extension.Type,
			Payload: dto.Extension.Payload,
		}
	}
	return echo
}
