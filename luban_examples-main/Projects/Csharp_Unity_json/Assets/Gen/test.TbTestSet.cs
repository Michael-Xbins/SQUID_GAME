
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
public partial class TbTestSet
{
    private readonly System.Collections.Generic.Dictionary<int, test.TestSet> _dataMap;
    private readonly System.Collections.Generic.List<test.TestSet> _dataList;
    
    public TbTestSet(JSONNode _buf)
    {
        _dataMap = new System.Collections.Generic.Dictionary<int, test.TestSet>();
        _dataList = new System.Collections.Generic.List<test.TestSet>();
        
        foreach(JSONNode _ele in _buf.Children)
        {
            test.TestSet _v;
            { if(!_ele.IsObject) { throw new SerializationException(); }  _v = test.TestSet.DeserializeTestSet(_ele);  }
            _dataList.Add(_v);
            _dataMap.Add(_v.Id, _v);
        }
    }

    public System.Collections.Generic.Dictionary<int, test.TestSet> DataMap => _dataMap;
    public System.Collections.Generic.List<test.TestSet> DataList => _dataList;

    public test.TestSet GetOrDefault(int key) => _dataMap.TryGetValue(key, out var v) ? v : null;
    public test.TestSet Get(int key) => _dataMap[key];
    public test.TestSet this[int key] => _dataMap[key];

    public void ResolveRef(Tables tables)
    {
        foreach(var _v in _dataList)
        {
            _v.ResolveRef(tables);
        }
    }

}

}
