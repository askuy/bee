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
	"unicode/utf8"
)

var DefaultBeegoPro = &Container{
	Option: Option{
		Dsn:          "",
		Driver:       "",
		ProType:      "",
		EnableModule: "",
		ApiPrefix:    "/",
		BeegoPath:    system.CurrentDir,
		Models:       make(map[string]string, 0),
		Url:          "git@github.com:beego-dev/bee-mod.git",
		Branch:       "master",
		GitPath:      system.BeegoHome + "/bee-mod",
		Overwrite:    false,
	},
	BeegoJson: system.CurrentDir + "/beegopro.json",
	CurPath:   system.CurrentDir,
	Render:    make(map[string]ProRenderMap, 0),
}

func init() {
	// 兼容默认的生成
	DefaultBeegoPro.Render["default"] = make(map[string]ProRender, 0)
	DefaultBeegoPro.Render["default"]["models"] = DefaultBeegoPro.renderModel
	DefaultBeegoPro.Render["default"]["controllers"] = DefaultBeegoPro.renderController

	// Ant Design后端 + 前端
	DefaultBeegoPro.Render["antDesign"] = make(map[string]ProRender, 0)
	DefaultBeegoPro.Render["antDesign"]["models"] = DefaultBeegoPro.renderModel
	DefaultBeegoPro.Render["antDesign"]["controllers"] = DefaultBeegoPro.renderController

	pongo2.RegisterFilter("lowerfirst", lwfirst)
	pongo2.RegisterFilter("upperfirst", upperfirst)
}

type Container struct {
	BeegoJson string
	Fields    string
	CurPath   string
	Option    Option
	Render    map[string]ProRenderMap
}

type Option struct {
	Dsn           string            `json:"dsn"`
	Driver        string            `json:"driver"`
	ProType       string            `json:"proType"`
	ApiPrefix     string            `json:"apiPrefix"`
	EnableModule  string            `json:"enableModule"`
	BeegoPath     string            `json:"beegoPath"`
	AntDesignPath string            `json:"antDesignPath"`
	Models        map[string]string `json:"models"`  // name => fields
	Url           string            `json:"url"`     // 安装路径
	Branch        string            `json:"branch"`  // 安装分支
	GitPath       string            `json:"gitPath"` // git clone隐藏地址
	Overwrite     bool              `json:"overwrite"`
}

type ProRender func(name, fields string) error // 渲染函数

type ProRenderMap map[string]ProRender // 渲染模板map

// Generate generates beego pro for a given path.
func (c *Container) Generate() {
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
	moduleMap, moduleFlag := c.Render[c.Option.ProType]
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

		// 找到需要的name和fields
		for name, fields := range c.Option.Models {
			err := render(name, fields)
			if err != nil {
				beeLogger.Log.Fatalf("beego pro render error, err: %s", err.Error())
			}
		}

	}

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
