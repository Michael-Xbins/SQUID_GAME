
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

using Luban;
using SimpleJSON;


namespace cfg.test
{
public partial class TbTestString
{
    private readonly System.Collections.Generic.Dictionary<int, test.TestString> _dataMap;
    private readonly System.Collections.Generic.List<test.TestString> _dataList;
    
    public TbTestString(JSONNode _buf)
    {
        _dataMap = new System.Collections.Generic.Dictionary<int, test.TestString>();
        _dataList = new System.Collections.Generic.List<test.TestString>();
        
        foreach(JSONNode _ele in _buf.Children)
        {
            test.TestString _v;
            { if(!_ele.IsObject) { throw new SerializationException(); }  _v = test.TestString.DeserializeTestString(_ele);  }
            _dataList.Add(_v);
            _dataMap.Add(_v.Id, _v);
        }
    }

    public System.Collections.Generic.Dictionary<int, test.TestString> DataMap => _dataMap;
    public System.Collections.Generic.List<test.TestString> DataList => _dataList;

    public test.TestString GetOrDefault(int key) => _dataMap.TryGetValue(key, out var v) ? v : null;
    public test.TestString Get(int key) => _dataMap[key];
    public test.TestString this[int key] => _dataMap[key];

    public void ResolveRef(Tables tables)
    {
        foreach(var _v in _dataList)
        {
            _v.ResolveRef(tables);
        }
    }

}

}
