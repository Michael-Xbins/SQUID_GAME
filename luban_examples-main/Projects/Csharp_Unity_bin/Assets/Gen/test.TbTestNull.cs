
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

using Luban;


namespace cfg.test
{
public partial class TbTestNull
{
    private readonly System.Collections.Generic.Dictionary<int, test.TestNull> _dataMap;
    private readonly System.Collections.Generic.List<test.TestNull> _dataList;
    
    public TbTestNull(ByteBuf _buf)
    {
        _dataMap = new System.Collections.Generic.Dictionary<int, test.TestNull>();
        _dataList = new System.Collections.Generic.List<test.TestNull>();
        
        for(int n = _buf.ReadSize() ; n > 0 ; --n)
        {
            test.TestNull _v;
            _v = test.TestNull.DeserializeTestNull(_buf);
            _dataList.Add(_v);
            _dataMap.Add(_v.Id, _v);
        }
    }

    public System.Collections.Generic.Dictionary<int, test.TestNull> DataMap => _dataMap;
    public System.Collections.Generic.List<test.TestNull> DataList => _dataList;

    public test.TestNull GetOrDefault(int key) => _dataMap.TryGetValue(key, out var v) ? v : null;
    public test.TestNull Get(int key) => _dataMap[key];
    public test.TestNull this[int key] => _dataMap[key];

    public void ResolveRef(Tables tables)
    {
        foreach(var _v in _dataList)
        {
            _v.ResolveRef(tables);
        }
    }

}

}
