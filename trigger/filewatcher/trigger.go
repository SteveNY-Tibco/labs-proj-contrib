/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package filewatcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
}

//-===============================-//
//     Define Trigger Factory
//-===============================-//

type Factory struct {
}

// Metadata implements trigger.Factory.Metadata
func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New implements trigger.Factory.New
func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	settings := &Settings{}
	err := metadata.MapToStruct(config.Settings, settings, true)
	if err != nil {
		return nil, err
	}

	return &Trigger{settings: settings}, nil
}

//-=========================-//
//      Define Trigger
//-=========================-//

var logger log.Logger

type Trigger struct {
	settings *Settings
	handlers []trigger.Handler
	mux      sync.Mutex
}

// Init implements trigger.Init
func (this *Trigger) Initialize(ctx trigger.InitContext) error {

	this.handlers = ctx.GetHandlers()
	logger = ctx.Logger()

	return nil
}

// Start implements ext.Trigger.Start
func (this *Trigger) Start() error {

	for handlerId, handler := range this.handlers {

		logger.Info("Start handler : name =  ", handler.Name())

		handlerSetting := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), handlerSetting, true)
		if err != nil {
			return err
		}

		if "" == handlerSetting.Foldername {
			return errors.New("Foldername not set yet!")
		}

		reader, err := NewFileWatcher(handlerId, handlerSetting.Foldername, handlerSetting.Filepattern, handlerSetting.CheckInterval)
		if err != nil {
			logger.Error("File reading error", err)
			return err
		}

		logger.Info("reader = ", reader)

		go reader.Start(this)

		if nil != err {
			return err
		}
	}

	return nil
}

// Stop implements ext.Trigger.Stop
func (t *Trigger) Stop() error {
	logger.Debug("Stopping endpoints")
	return nil
}

func (t *Trigger) HandleContent(handlerId int, id string, path string, filename string, modifiedTime int64, checkTime int64) {
	t.mux.Lock()
	defer t.mux.Unlock()
	outputData := &Output{}
	outputData.MessageID = id
	outputData.Filepath = path
	outputData.Filename = filename
	outputData.ModifiedTime = modifiedTime
	outputData.CheckTime = checkTime

	logger.Info("(FileContentHandler.HandleContent) - Trigger sends MessageId : ", id, ", handlerId : ", handlerId, ", path : ", path, ", filename : ", filename, ", modifiedTime : ", modifiedTime)

	_, err := t.handlers[handlerId].Handle(context.Background(), outputData)

	if nil != err {
		logger.Errorf("Run action for handler [%s] failed for reason [%s] message lost", t.handlers[handlerId], err)
	}
}

//-====================-//
//    File Watcher
//-====================-//

func NewFileWatcher(handlerId int, foldername string, filepattern string, checkInteval int) (FileWatcher, error) {
	var fileWatcher FileWatcher
	if stat, err := os.Stat(foldername); err == nil {
		switch mode := stat.Mode(); {
		case mode.IsDir():
			fmt.Println("directory")
			fileWatcher = FolderReader{
				foldername:   foldername,
				handlerId:    handlerId,
				filepattern:  filepattern,
				checkInteval: checkInteval,
			}
		case mode.IsRegular():
			fmt.Println("file")
			return nil, errors.New("Not a folder !!!!!!!!!")
		}
	} else if os.IsNotExist(err) {
		return nil, err

	} else {
		return nil, err
	}
	return fileWatcher, nil
}

type FileContentHandler interface {
	HandleContent(handlerId int, id string, filepath string, filename string, time int64, checkTime int64)
}

type FileWatcher interface {
	Start(handler FileContentHandler) error
}

type FolderReader struct {
	handlerId    int
	foldername   string
	filepattern  string
	checkInteval int
}

func (this FolderReader) Start(handler FileContentHandler) error {

	//	checkInterval = this.checkInteval
	libRegEx, err := regexp.Compile(this.filepattern)
	if err != nil {
		logger.Error(err)
	}

	for {
		time.Sleep(time.Duration(this.checkInteval) * time.Second)
		err = filepath.Walk(this.foldername, func(path string, info os.FileInfo, err error) error {
			if err == nil && false == info.IsDir() && libRegEx.MatchString(info.Name()) {
				handler.HandleContent(this.handlerId, "messageid", path, info.Name(), info.ModTime().Unix(), time.Now().Unix())
			}
			return nil
		})
		if err != nil {
			logger.Error(err)
			break
		}
	}

	return err
}
