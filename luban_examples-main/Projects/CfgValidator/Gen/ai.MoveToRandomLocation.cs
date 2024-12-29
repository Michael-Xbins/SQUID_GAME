
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

using Luban;
using System.Text.Json;


namespace cfg.ai
{
public sealed partial class MoveToRandomLocation : ai.Task
{
    public MoveToRandomLocation(JsonElement _buf)  : base(_buf) 
    {
        OriginPositionKey = _buf.GetProperty("origin_position_key").GetString();
        Radius = _buf.GetProperty("radius").GetSingle();
    }

    public static MoveToRandomLocation DeserializeMoveToRandomLocation(JsonElement _buf)
    {
        return new ai.MoveToRandomLocation(_buf);
    }

    public readonly string OriginPositionKey;
    public readonly float Radius;
   
    public const int __ID__ = -2140042998;
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
        + "decorators:" + Luban.StringUtil.CollectionToString(Decorators) + ","
        + "services:" + Luban.StringUtil.CollectionToString(Services) + ","
        + "ignoreRestartSelf:" + IgnoreRestartSelf + ","
        + "originPositionKey:" + OriginPositionKey + ","
        + "radius:" + Radius + ","
        + "}";
    }
}

}
