
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

using Luban;


namespace cfg.tag
{
public sealed partial class TestTag : Luban.BeanBase
{
    public TestTag(ByteBuf _buf) 
    {
        Id = _buf.ReadInt();
        Value = _buf.ReadString();
    }

    public static TestTag DeserializeTestTag(ByteBuf _buf)
    {
        return new tag.TestTag(_buf);
    }

    public readonly int Id;
    public readonly string Value;
   
    public const int __ID__ = 1742933812;
    public override int GetTypeId() => __ID__;

    public  void ResolveRef(Tables tables)
    {
        
        
    }

    public override string ToString()
    {
        return "{ "
        + "id:" + Id + ","
        + "value:" + Value + ","
        + "}";
    }
}

}
