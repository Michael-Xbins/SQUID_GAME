
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type Compete struct {
    Id string
    Name string
    Desc string
    NumInt int32
}

const TypeId_Compete = -1679825913

func (*Compete) GetTypeId() int32 {
    return -1679825913
}

func NewCompete(_buf map[string]interface{}) (_v *Compete, err error) {
    _v = &Compete{}
    { var _ok_ bool; if _v.Id, _ok_ = _buf["id"].(string); !_ok_ { err = errors.New("id error"); return } }
    { var _ok_ bool; if _v.Name, _ok_ = _buf["name"].(string); !_ok_ { err = errors.New("name error"); return } }
    { var _ok_ bool; if _v.Desc, _ok_ = _buf["desc"].(string); !_ok_ { err = errors.New("desc error"); return } }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["num_int"].(float64); !_ok_ { err = errors.New("num_int error"); return }; _v.NumInt = int32(_tempNum_) }
    return
}

