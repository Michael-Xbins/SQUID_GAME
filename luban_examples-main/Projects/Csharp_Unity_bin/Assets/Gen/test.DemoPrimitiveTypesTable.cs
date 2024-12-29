
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
public sealed partial class DemoPrimitiveTypesTable : Luban.BeanBase
{
    public DemoPrimitiveTypesTable(ByteBuf _buf) 
    {
        X1 = _buf.ReadBool();
        X2 = _buf.ReadByte();
        X3 = _buf.ReadShort();
        X4 = _buf.ReadInt();
        X5 = _buf.ReadLong();
        X6 = _buf.ReadFloat();
        X7 = _buf.ReadDouble();
        S1 = _buf.ReadString();
        S2 = _buf.ReadString();
        V2 = vector2.Deserializevector2(_buf);
        V3 = vector3.Deserializevector3(_buf);
        V4 = vector4.Deserializevector4(_buf);
        T1 = _buf.ReadLong();
    }

    public static DemoPrimitiveTypesTable DeserializeDemoPrimitiveTypesTable(ByteBuf _buf)
    {
        return new test.DemoPrimitiveTypesTable(_buf);
    }

    public readonly bool X1;
    public readonly byte X2;
    public readonly short X3;
    public readonly int X4;
    public readonly long X5;
    public readonly float X6;
    public readonly double X7;
    public readonly string S1;
    public readonly string S2;
    public readonly vector2 V2;
    public readonly vector3 V3;
    public readonly vector4 V4;
    public readonly long T1;
   
    public const int __ID__ = -370934083;
    public override int GetTypeId() => __ID__;

    public  void ResolveRef(Tables tables)
    {
        
        
        
        
        
        
        
        
        
        
        
        
        
    }

    public override string ToString()
    {
        return "{ "
        + "x1:" + X1 + ","
        + "x2:" + X2 + ","
        + "x3:" + X3 + ","
        + "x4:" + X4 + ","
        + "x5:" + X5 + ","
        + "x6:" + X6 + ","
        + "x7:" + X7 + ","
        + "s1:" + S1 + ","
        + "s2:" + S2 + ","
        + "v2:" + V2 + ","
        + "v3:" + V3 + ","
        + "v4:" + V4 + ","
        + "t1:" + T1 + ","
        + "}";
    }
}

}
