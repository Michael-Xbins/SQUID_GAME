
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
public sealed partial class ChooseTarget : ai.Service
{
    public ChooseTarget(JSONNode _buf)  : base(_buf) 
    {
        { if(!_buf["result_target_key"].IsString) { throw new SerializationException(); }  ResultTargetKey = _buf["result_target_key"]; }
    }

    public static ChooseTarget DeserializeChooseTarget(JSONNode _buf)
    {
        return new ai.ChooseTarget(_buf);
    }

    public readonly string ResultTargetKey;
   
    public const int __ID__ = 1601247918;
    public override int GetTypeId() => __ID__;

    public override void ResolveRef(Tables tables)
    {
        base.ResolveRef(tables);
        
    }

    public override string ToString()
    {
        return "{ "
        + "id:" + Id + ","
        + "nodeName:" + NodeName + ","
        + "resultTargetKey:" + ResultTargetKey + ","
        + "}";
    }
}

}
