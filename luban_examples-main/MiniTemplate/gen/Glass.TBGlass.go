
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type GlassTBGlass struct {
    _dataMap map[string]*Glass
    _dataList []*Glass
}

func NewGlassTBGlass(_buf []map[string]interface{}) (*GlassTBGlass, error) {
    _dataList := make([]*Glass, 0, len(_buf))
    dataMap := make(map[string]*Glass)

    for _, _ele_ := range _buf {
        if _v, err2 := NewGlass(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.Id] = _v
        }
    }
    return &GlassTBGlass{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *GlassTBGlass) GetDataMap() map[string]*Glass {
    return table._dataMap
}

func (table *GlassTBGlass) GetDataList() []*Glass {
    return table._dataList
}

func (table *GlassTBGlass) Get(key string) *Glass {
    return table._dataMap[key]
}


