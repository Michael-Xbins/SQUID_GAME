
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg.test;

import luban.*;


public final class DemoGroup extends AbstractBean {
    public DemoGroup(ByteBuf _buf) { 
        id = _buf.readInt();
        x1 = _buf.readInt();
        x2 = _buf.readInt();
        x3 = _buf.readInt();
        x4 = _buf.readInt();
        x5 = cfg.test.InnerGroup.deserialize(_buf);
    }

    public static DemoGroup deserialize(ByteBuf _buf) {
            return new cfg.test.DemoGroup(_buf);
    }

    public final int id;
    public final int x1;
    public final int x2;
    public final int x3;
    public final int x4;
    public final cfg.test.InnerGroup x5;

    public static final int __ID__ = -379263008;
    
    @Override
    public int getTypeId() { return __ID__; }

    @Override
    public String toString() {
        return "{ "
        + "(format_field_name __code_style field.name):" + id + ","
        + "(format_field_name __code_style field.name):" + x1 + ","
        + "(format_field_name __code_style field.name):" + x2 + ","
        + "(format_field_name __code_style field.name):" + x3 + ","
        + "(format_field_name __code_style field.name):" + x4 + ","
        + "(format_field_name __code_style field.name):" + x5 + ","
        + "}";
    }
}

