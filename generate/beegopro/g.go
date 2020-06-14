// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package beegopro

import (
	"errors"
	"fmt"
	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/utils"
	"github.com/flosch/pongo2"
	"github.com/smartwalle/pongo2render"
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

var SQLDriver utils.DocValue
var Level utils.DocValue
var Tables utils.DocValue
var Fields utils.DocValue
var DDL utils.DocValue

// render
type GenRender struct {
	Context pongo2.Context
	Option  Option
	*pongo2render.Render
	Name        string
	PackageName string
	FlushFile   string
	PkgPath     string
}

func NewGenRender(packageName string, name string, option Option) *GenRender {
	p, f := path.Split(name)
	title := strings.Title(f)

	obj := &GenRender{
		Context:     make(pongo2.Context, 0),
		Option:      option,
		Name:        title,
		PackageName: packageName,
	}
	// render path
	obj.Render = pongo2render.NewRender(option.GitPath + "/" + option.ProType + "/" + option.ProVersion + "/" + obj.PackageName)

	if p != "" {
		i := strings.LastIndex(p[:len(p)-1], "/")
		packageName = p[i+1 : len(p)-1]
	}

	beeLogger.Log.Infof("Using '%s' as name from %s", title, obj.PackageName)
	beeLogger.Log.Infof("Using '%s' as package name from %s", packageName, obj.PackageName)

	fp := path.Join(obj.Option.BeegoPath, packageName, p)
	err := createPath(fp)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create the controllers directory: %s", err)
	}

	obj.FlushFile = path.Join(fp, strings.ToLower(title)+".go")
	obj.PkgPath = getPackagePath(obj.Option.BeegoPath)

	obj.Context["packageName"] = obj.PackageName
	obj.Context["name"] = obj.Name
	obj.Context["pkgPath"] = obj.PkgPath
	obj.Context["apiPrefix"] = obj.Option.ApiPrefix
	return obj
}

func (r *GenRender) SetContext(key string, value interface{}) {
	r.Context[key] = value
}

func (r *GenRender) Exec(name string) {
	var (
		buf string
		err error
	)
	buf, err = r.Render.Template(name).Execute(r.Context)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create the %s render tmpl: %s", name, err)
		return
	}
	err = r.write(r.FlushFile, buf)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create file: %s", err)
		return
	}
	beeLogger.Log.Infof("create file '%s' from %s", r.FlushFile, r.PackageName)
}

// write 写bytes到文件
func (c *GenRender) write(filename string, buf string) (err error) {
	if !c.Option.Overwrite && utils.IsExist(filename) {
		err = errors.New("file is exist, path is " + filename)
		return
	}

	if c.Option.Overwrite && utils.IsExist(filename) {
		bakName := fmt.Sprintf("%s.%s.bak", filename, time.Now().Format("2006.01.02.15.04.05"))
		beeLogger.Log.Infof("bak file '%s'", bakName)
		if err := os.Rename(filename, bakName); err != nil {
			err = errors.New("file is bak error, path is " + bakName)
			return err
		}
	}

	filePath := path.Dir(filename)
	err = createPath(filePath)
	if err != nil {
		err = errors.New("write create path " + err.Error())
		return
	}

	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		err = errors.New("write create file " + err.Error())
		return
	}

	// 格式化代码
	bts, err := format.Source([]byte(buf))
	if err != nil {
		err = errors.New("format buf error " + err.Error())
		return
	}

	err = ioutil.WriteFile(filename, bts, 0644)
	if err != nil {
		err = errors.New("write write file " + err.Error())
		return
	}
	return
}
