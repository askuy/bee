package beegopro

import (
	"encoding/json"
	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/pkg/system"
	"github.com/beego/bee/utils"
	"github.com/flosch/pongo2"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"
)

var DefaultBeegoPro = &Container{
	Option: Option{
		Dsn:           "",
		Driver:        "",
		ProType:       "default",
		ProVersion:    "v1",
		EnableModule:  "",
		ApiPrefix:     "/",
		BeegoPath:     system.CurrentDir,
		AntDesignPath: system.CurrentDir,
		Models:        make(map[string]ModelsContent, 0),
		Url:           "git@github.com:beego-dev/bee-mod.git",
		Branch:        "master",
		GitPath:       system.BeegoHome + "/bee-mod",
		Overwrite:     false,
	},
	BeegoJson:      system.CurrentDir + "/beegopro.json",
	CurPath:        system.CurrentDir,
	SingleRender:   make(map[string]ProSingleRenderMap, 0),
	GlobalRender:   make(map[string]ProGlobalRenderMap, 0),
	ControllerOnce: &sync.Once{},
	ModelOnce:      &sync.Once{},
}

func init() {
	// 兼容默认的生成
	DefaultBeegoPro.SingleRender["default"] = make(map[string]ProSingleRender, 0)
	DefaultBeegoPro.SingleRender["default"]["models"] = DefaultBeegoPro.renderModel
	DefaultBeegoPro.SingleRender["default"]["controllers"] = DefaultBeegoPro.renderController
	DefaultBeegoPro.SingleRender["default"]["routers"] = DefaultBeegoPro.renderRouter

	//DefaultBeegoPro.GlobalRender["default"] = make(map[string]ProGlobalRender, 0)
	//DefaultBeegoPro.GlobalRender["default"]["routers"] = DefaultBeegoPro.renderRouter

	// Ant Design后端 + 前端
	DefaultBeegoPro.SingleRender["antDesign"] = make(map[string]ProSingleRender, 0)
	DefaultBeegoPro.SingleRender["antDesign"]["models"] = DefaultBeegoPro.renderModel
	DefaultBeegoPro.SingleRender["antDesign"]["controllers"] = DefaultBeegoPro.renderController
	DefaultBeegoPro.SingleRender["antDesign"]["routers"] = DefaultBeegoPro.renderRouter
	DefaultBeegoPro.SingleRender["antDesign"]["migrations"] = DefaultBeegoPro.renderMigration
	DefaultBeegoPro.SingleRender["antDesign"]["antList"] = DefaultBeegoPro.renderAntList
	DefaultBeegoPro.SingleRender["antDesign"]["antForm"] = DefaultBeegoPro.renderAntForm

	pongo2.RegisterFilter("lowerfirst", lwfirst)
	pongo2.RegisterFilter("upperfirst", upperfirst)
}

type Container struct {
	BeegoJson      string
	Fields         string
	CurPath        string
	Option         Option
	SingleRender   map[string]ProSingleRenderMap
	GlobalRender   map[string]ProGlobalRenderMap
	ControllerOnce *sync.Once
	ModelOnce      *sync.Once
}

type Option struct {
	Dsn           string                   `json:"dsn"`
	Driver        string                   `json:"driver"`
	ProType       string                   `json:"proType"`
	ProVersion    string                   `json:"proVersion"`
	ApiPrefix     string                   `json:"apiPrefix"`
	EnableModule  string                   `json:"enableModule"`
	BeegoPath     string                   `json:"beegoPath"`
	AntDesignPath string                   `json:"antDesignPath"`
	Models        map[string]ModelsContent `json:"models"`  // name => fields
	Url           string                   `json:"url"`     // 安装路径
	Branch        string                   `json:"branch"`  // 安装分支
	GitPath       string                   `json:"gitPath"` // git clone隐藏地址
	Overwrite     bool                     `json:"overwrite"`
}

type ModelsContent struct {
	Schema    []Schema `json:"schema"`
	SourceGen string   `json:"sourceGen"`
	ApiPrefix string   `json:"apiPrefix"`
}

type Schema struct {
	Name    string `json:"name"`    // mysql name
	Type    string `json:"type"`    // mysql type
	Comment string `json:"comment"` // mysql comment
}

type ProSingleRender func(name string, content ModelsContent) error // 渲染单个表
type ProGlobalRender func() error                                   // 渲染单个表

type ProSingleRenderMap map[string]ProSingleRender // 渲染模板map
type ProGlobalRenderMap map[string]ProGlobalRender // 渲染模板map

// Generate generates beego pro for a given path.
func (c *Container) Generate(flag bool) {
	//
	if flag {
		if !utils.IsExist(c.BeegoJson) {
			beeLogger.Log.Fatalf("beego pro json is not exist, beego json path: %s", c.BeegoJson)
			return
		}

		content, err := ioutil.ReadFile(c.BeegoJson)
		if err != nil {
			beeLogger.Log.Fatalf("read beego pro error, err: %s", err.Error())
			return
		}
		err = json.Unmarshal(content, &c.Option)
		if err != nil {
			beeLogger.Log.Fatalf("beego json unmarshal error, err: %s", err.Error())
			return
		}
	}

	absolutePath, err := filepath.Abs(c.Option.BeegoPath)
	if err != nil {
		beeLogger.Log.Fatalf("beego pro beego path error, err: %s", err.Error())
		return
	}

	c.Option.BeegoPath = absolutePath

	//err = git.CloneORPullRepo(c.Option.Url, c.Option.GitPath)
	//if err != nil {
	//	beeLogger.Log.Fatalf("beego pro git clone or pull repo error, err: %s", err)
	//	return
	//}
	c.render()
}

func (c *Container) render() {
	arr := strings.Split(c.Option.EnableModule, ",")
	moduleMap, moduleFlag := c.SingleRender[c.Option.ProType]
	if !moduleFlag {
		beeLogger.Log.Fatalf("beego json pro type not exist, pro type is: %s", c.Option.ProType)
		return
	}

	for _, moduleName := range arr {
		// 找到渲染函数
		render, flag := moduleMap[moduleName]
		if !flag {
			continue
		}

		// 找到需要的table name和fields
		for name, content := range c.Option.Models {
			err := render(name, content)
			if err != nil {
				beeLogger.Log.Fatalf("beego pro render error, err: %s", err.Error())
			}
		}
	}

	// global render
	//globalModuleMap, globalModuleFlag := c.GlobalRender[c.Option.ProType]
	//if !globalModuleFlag {
	//	beeLogger.Log.Fatalf("beego json pro global type not exist, pro type is: %s", c.Option.ProType)
	//	return
	//}
	//
	//for _, moduleName := range arr {
	//	// 找到渲染函数
	//	globalRender, flag := globalModuleMap[moduleName]
	//	if !flag {
	//		continue
	//	}
	//
	//	err := globalRender()
	//	if err != nil {
	//		beeLogger.Log.Fatalf("beego pro render error, err: %s", err.Error())
	//	}
	//}

}

// lwfirst 首字母小写，注意不要和go关键字冲突
func lwfirst(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.Len() <= 0 {
		return pongo2.AsValue(""), nil
	}
	t := in.String()
	r, size := utf8.DecodeRuneInString(t)
	return pongo2.AsValue(strings.ToLower(string(r)) + t[size:]), nil
}

func upperfirst(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.Len() <= 0 {
		return pongo2.AsValue(""), nil
	}
	t := in.String()
	return pongo2.AsValue(strings.Replace(t, string(t[0]), strings.ToUpper(string(t[0])), 1)), nil
}
