package filewatcher

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
}
type HandlerSettings struct {
	Foldername    string `md:"Foldername"`
	Filepattern   string `md:"EmitPerLine"`
	CheckInterval int    `md:"CheckInterval"`
}

type Output struct {
	MessageID    string `md:"MessageID"`
	Filepath     string `md:"Filepath"`
	Filename     string `md:"Filename"`
	ModifiedTime int64  `md:"ModifiedTime"`
	CheckTime    int64  `md:"CheckTime"`
}

func (this *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"MessageID":    this.MessageID,
		"Filepath":     this.Filepath,
		"Filename":     this.Filename,
		"ModifiedTime": this.ModifiedTime,
		"CheckTime":    this.CheckTime,
	}
}

func (this *Output) FromMap(values map[string]interface{}) error {

	var err error
	this.MessageID, err = coerce.ToString(values["MessageID"])
	this.Filepath, err = coerce.ToString(values["Filepath"])
	this.Filename, err = coerce.ToString(values["Filename"])
	this.ModifiedTime, err = coerce.ToInt64(values["ModifiedTime"])
	this.CheckTime, err = coerce.ToInt64(values["CheckTime"])
	if err != nil {
		return err
	}

	return nil
}
