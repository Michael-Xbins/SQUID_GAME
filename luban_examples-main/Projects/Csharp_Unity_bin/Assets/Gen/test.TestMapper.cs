
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
public sealed partial class TestMapper : Luban.BeanBase
{
    public TestMapper(ByteBuf _buf) 
    {
        Id = _buf.ReadInt();
        AudioType = (AudioType)_buf.ReadInt();
        V2 = vector2.Deserializevector2(_buf);
    }

    public static TestMapper DeserializeTestMapper(ByteBuf _buf)
    {
        return new test.TestMapper(_buf);
    }

    public readonly int Id;
    public readonly AudioType AudioType;
    public readonly vector2 V2;
   
    public const int __ID__ = 149110895;
    public override int GetTypeId() => __ID__;

    public  void ResolveRef(Tables tables)
    {
        
        
        
    }

    public override string ToString()
    {
        return "{ "
        + "id:" + Id + ","
        + "audioType:" + AudioType + ","
        + "v2:" + V2 + ","
        + "}";
    }
}

}
