
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg.test;

import luban.*;


public final class InnerGroup extends AbstractBean {
    public InnerGroup(ByteBuf _buf) { 
        y1 = _buf.readInt();
        y2 = _buf.readInt();
        y3 = _buf.readInt();
        y4 = _buf.readInt();
    }

    public static InnerGroup deserialize(ByteBuf _buf) {
            return new cfg.test.InnerGroup(_buf);
    }

    public final int y1;
    public final int y2;
    public final int y3;
    public final int y4;

    public static final int __ID__ = -587873083;
    
    @Override
    public int getTypeId() { return __ID__; }

    @Override
    public String toString() {
        return "{ "
        + "(format_field_name __code_style field.name):" + y1 + ","
        + "(format_field_name __code_style field.name):" + y2 + ","
        + "(format_field_name __code_style field.name):" + y3 + ","
        + "(format_field_name __code_style field.name):" + y4 + ","
        + "}";
    }
}

