
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
public abstract partial class RefDynamicBase : Luban.BeanBase
{
    public RefDynamicBase(JSONNode _buf) 
    {
        { if(!_buf["x"].IsNumber) { throw new SerializationException(); }  X = _buf["x"]; }
        X_Ref = null;
    }

    public static RefDynamicBase DeserializeRefDynamicBase(JSONNode _buf)
    {
        switch ((string)_buf["$type"])
        {
            case "RefBean": return new test.RefBean(_buf);
            default: throw new SerializationException();
        }
    }

    public readonly int X;
    public test.TestBeRef X_Ref;
   

    public virtual void ResolveRef(Tables tables)
    {
        X_Ref = tables.TbTestBeRef.GetOrDefault(X);
    }

    public override string ToString()
    {
        return "{ "
        + "x:" + X + ","
        + "}";
    }
}

}
