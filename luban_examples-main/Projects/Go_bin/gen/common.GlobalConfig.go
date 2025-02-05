
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

type CommonGlobalConfig struct {
    BagCapacity int32
    BagCapacitySpecial int32
    BagTempExpendableCapacity int32
    BagTempToolCapacity int32
    BagInitCapacity int32
    QuickBagCapacity int32
    ClothBagCapacity int32
    ClothBagInitCapacity int32
    ClothBagCapacitySpecial int32
    BagInitItemsDropId *int32
    MailBoxCapacity int32
    DamageParamC float32
    DamageParamE float32
    DamageParamF float32
    DamageParamD float32
    RoleSpeed float32
    MonsterSpeed float32
    InitEnergy int32
    InitViality int32
    MaxViality int32
    PerVialityRecoveryTime int32
}

const TypeId_CommonGlobalConfig = -848234488

func (*CommonGlobalConfig) GetTypeId() int32 {
    return -848234488
}

func NewCommonGlobalConfig(_buf *luban.ByteBuf) (_v *CommonGlobalConfig, err error) {
    _v = &CommonGlobalConfig{}
    { if _v.BagCapacity, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.BagCapacitySpecial, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.BagTempExpendableCapacity, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.BagTempToolCapacity, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.BagInitCapacity, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.QuickBagCapacity, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.ClothBagCapacity, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.ClothBagInitCapacity, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.ClothBagCapacitySpecial, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { var __exists__ bool; if __exists__, err = _buf.ReadBool(); err != nil { return }; if __exists__ { var __x__ int32;  { if __x__, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }; _v.BagInitItemsDropId = &__x__ }}
    { if _v.MailBoxCapacity, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.DamageParamC, err = _buf.ReadFloat(); err != nil { err = errors.New("error"); return } }
    { if _v.DamageParamE, err = _buf.ReadFloat(); err != nil { err = errors.New("error"); return } }
    { if _v.DamageParamF, err = _buf.ReadFloat(); err != nil { err = errors.New("error"); return } }
    { if _v.DamageParamD, err = _buf.ReadFloat(); err != nil { err = errors.New("error"); return } }
    { if _v.RoleSpeed, err = _buf.ReadFloat(); err != nil { err = errors.New("error"); return } }
    { if _v.MonsterSpeed, err = _buf.ReadFloat(); err != nil { err = errors.New("error"); return } }
    { if _v.InitEnergy, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.InitViality, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.MaxViality, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    { if _v.PerVialityRecoveryTime, err = _buf.ReadInt(); err != nil { err = errors.New("error"); return } }
    return
}

