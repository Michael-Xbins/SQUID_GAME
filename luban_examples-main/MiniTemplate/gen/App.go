
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type App struct {
    Id string
    Name string
    Target int32
    Desc string
    NumInt int32
}

const TypeId_App = 66049

func (*App) GetTypeId() int32 {
    return 66049
}

func NewApp(_buf map[string]interface{}) (_v *App, err error) {
    _v = &App{}
    { var _ok_ bool; if _v.Id, _ok_ = _buf["id"].(string); !_ok_ { err = errors.New("id error"); return } }
    { var _ok_ bool; if _v.Name, _ok_ = _buf["name"].(string); !_ok_ { err = errors.New("name error"); return } }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["target"].(float64); !_ok_ { err = errors.New("target error"); return }; _v.Target = int32(_tempNum_) }
    { var _ok_ bool; if _v.Desc, _ok_ = _buf["desc"].(string); !_ok_ { err = errors.New("desc error"); return } }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["num_int"].(float64); !_ok_ { err = errors.New("num_int error"); return }; _v.NumInt = int32(_tempNum_) }
    return
}
