
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg.test;

import luban.*;


public final class H1 extends AbstractBean {
    public H1(ByteBuf _buf) { 
        y2 = cfg.test.H2.deserialize(_buf);
        y3 = _buf.readInt();
    }

    public static H1 deserialize(ByteBuf _buf) {
            return new cfg.test.H1(_buf);
    }

    public final cfg.test.H2 y2;
    public final int y3;

    public static final int __ID__ = -1422503995;
    
    @Override
    public int getTypeId() { return __ID__; }

    @Override
    public String toString() {
        return "{ "
        + "(format_field_name __code_style field.name):" + y2 + ","
        + "(format_field_name __code_style field.name):" + y3 + ","
        + "}";
    }
}

