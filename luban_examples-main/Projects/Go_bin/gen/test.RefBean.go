
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import (
    "demo/luban"
)

import "errors"

type TestRefBean struct {
    X int32
    Arr []int32
}

const TypeId_TestRefBean = 1963260263

func (*TestRefBean) GetTypeId() int32 {
    return 1963260263
}

func NewTestRefBean(_buf *luban.ByteBuf) (_v *TestRefBean, err error) {
    _v = &TestRefBean{}
    { if _v.X, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    {_v.Arr = make([]int32, 0); var _n_ int; if _n_, err = _buf.ReadSize(); err != nil { err = errors.New("error"); return}; for i := 0 ; i < _n_ ; i++ { var _e_ int32; { if _e_, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }; _v.Arr = append(_v.Arr, _e_) } }
    return
}
