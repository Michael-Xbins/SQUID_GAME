
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

type TestTestIndex struct {
    Id int32
    Eles []*TestDemoType1
}

const TypeId_TestTestIndex = 1941154020

func (*TestTestIndex) GetTypeId() int32 {
    return 1941154020
}

func NewTestTestIndex(_buf *luban.ByteBuf) (_v *TestTestIndex, err error) {
    _v = &TestTestIndex{}
    { if _v.Id, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    {_v.Eles = make([]*TestDemoType1, 0); var _n_ int; if _n_, err = _buf.ReadSize(); err != nil { err = errors.New("error"); return}; for i := 0 ; i < _n_ ; i++ { var _e_ *TestDemoType1; { if _e_, err = NewTestDemoType1(_buf); err != nil { err = errors.New("error"); return } }; _v.Eles = append(_v.Eles, _e_) } }
    return
}
