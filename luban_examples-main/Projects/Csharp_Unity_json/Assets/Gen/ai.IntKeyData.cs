
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

using Luban;
using SimpleJSON;


namespace cfg.ai
{
public sealed partial class IntKeyData : ai.KeyData
{
    public IntKeyData(JSONNode _buf)  : base(_buf) 
    {
        { if(!_buf["value"].IsNumber) { throw new SerializationException(); }  Value = _buf["value"]; }
    }

    public static IntKeyData DeserializeIntKeyData(JSONNode _buf)
    {
        return new ai.IntKeyData(_buf);
    }

    public readonly int Value;
   
    public const int __ID__ = -342751904;
    public override int GetTypeId() => __ID__;

    public override void ResolveRef(Tables tables)
    {
        base.ResolveRef(tables);
        
    }

    public override string ToString()
    {
        return "{ "
        + "value:" + Value + ","
        + "}";
    }
}

}
