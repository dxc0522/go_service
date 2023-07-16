since find the version( v0.9.29 )  has some logic error and can't achieve our purpose

so adjust some code,mainly as following:

+ make lint passed changes
+ remove some unuse logic code
+ change the functions form : from package to the Option's method

**specially**

+ in file `tool.go`

  add code `modelInfo["UseGuregu"] = opt.Guregu` in `func (opt *Option) Generate(conf *dbmeta.Config) error`
  to solve err on  render `model.go.tmpl` at   `{{if .UseGuregu}} "github.com/guregu/null" {{end}}`

  Err msg

  ```
  error in rendering internal://model.go.tmpl: template: model.go.tmpl:8:9: executing "model.go.tmpl" at <.UseGuregu>: map has no entry for key "UseGuregu"
  ```

+ in file  `model.go.tmpl`

  Add

  ```
  var (
  	_ = time.Second
  	_ = sql.LevelDefault
  	 {{if .UseGuregu}} _ = null.Bool{} {{end}}
  	_ = uuid.UUID{}
  )
  ```

  REMOVE

    + The `/* */` arround `type {{.StructName}} struct {`




