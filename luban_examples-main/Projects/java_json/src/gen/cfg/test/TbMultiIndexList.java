
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg.test;

import luban.*;
import com.google.gson.JsonElement;


public final class TbMultiIndexList {
    private final java.util.ArrayList<cfg.test.MultiIndexList> _dataList;
    
    public TbMultiIndexList(JsonElement _buf) {
        _dataList = new java.util.ArrayList<cfg.test.MultiIndexList>();
        
        for (com.google.gson.JsonElement _e_ : _buf.getAsJsonArray()) {
            cfg.test.MultiIndexList _v;
            _v = cfg.test.MultiIndexList.deserialize(_e_.getAsJsonObject());
            _dataList.add(_v);
        }
    }

    public java.util.ArrayList<cfg.test.MultiIndexList> getDataList() { return _dataList; }

    public cfg.test.MultiIndexList get(int index) { return _dataList.get(index); }


}
