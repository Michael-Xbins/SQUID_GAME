
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;

import luban.*;


public final class DefineFromExcel2 extends AbstractBean {
    public DefineFromExcel2(ByteBuf _buf) { 
        id = _buf.readInt();
        x1 = _buf.readBool();
        x5 = _buf.readLong();
        x6 = _buf.readFloat();
        x8 = _buf.readInt();
        x10 = _buf.readString();
        x13 = _buf.readInt();
        x132 = _buf.readInt();
        x14 = cfg.test.DemoDynamic.deserialize(_buf);
        x15 = cfg.test.Shape.deserialize(_buf);
        v2 = cfg.vector2.deserialize(_buf);
        t1 = _buf.readLong();
        {int n = Math.min(_buf.readSize(), _buf.size());k1 = new int[n];for(int i = 0 ; i < n ; i++) { int _e;_e = _buf.readInt(); k1[i] = _e;}}
        {int n = Math.min(_buf.readSize(), _buf.size());k2 = new int[n];for(int i = 0 ; i < n ; i++) { int _e;_e = _buf.readInt(); k2[i] = _e;}}
        {int n = Math.min(_buf.readSize(), _buf.size());k8 = new java.util.HashMap<Integer, Integer>(n * 3 / 2);for(int i = 0 ; i < n ; i++) { Integer _k;  _k = _buf.readInt(); Integer _v;  _v = _buf.readInt();     k8.put(_k, _v);}}
        {int n = Math.min(_buf.readSize(), _buf.size());k9 = new java.util.ArrayList<cfg.test.DemoE2>(n);for(int i = 0 ; i < n ; i++) { cfg.test.DemoE2 _e;  _e = cfg.test.DemoE2.deserialize(_buf); k9.add(_e);}}
        {int n = Math.min(_buf.readSize(), _buf.size());k10 = new java.util.ArrayList<cfg.vector3>(n);for(int i = 0 ; i < n ; i++) { cfg.vector3 _e;  _e = cfg.vector3.deserialize(_buf); k10.add(_e);}}
        {int n = Math.min(_buf.readSize(), _buf.size());k11 = new java.util.ArrayList<cfg.vector4>(n);for(int i = 0 ; i < n ; i++) { cfg.vector4 _e;  _e = cfg.vector4.deserialize(_buf); k11.add(_e);}}
    }

    public static DefineFromExcel2 deserialize(ByteBuf _buf) {
            return new cfg.DefineFromExcel2(_buf);
    }

    /**
     * 这是id
     */
    public final int id;
    /**
     * 字段x1
     */
    public final boolean x1;
    public final long x5;
    public final float x6;
    public final int x8;
    public final String x10;
    public final int x13;
    public final int x132;
    public final cfg.test.DemoDynamic x14;
    public final cfg.test.Shape x15;
    public final cfg.vector2 v2;
    public final long t1;
    public final int[] k1;
    public final int[] k2;
    public final java.util.HashMap<Integer, Integer> k8;
    public final java.util.ArrayList<cfg.test.DemoE2> k9;
    public final java.util.ArrayList<cfg.vector3> k10;
    public final java.util.ArrayList<cfg.vector4> k11;

    public static final int __ID__ = 482045152;
    
    @Override
    public int getTypeId() { return __ID__; }

    @Override
    public String toString() {
        return "{ "
        + "(format_field_name __code_style field.name):" + id + ","
        + "(format_field_name __code_style field.name):" + x1 + ","
        + "(format_field_name __code_style field.name):" + x5 + ","
        + "(format_field_name __code_style field.name):" + x6 + ","
        + "(format_field_name __code_style field.name):" + x8 + ","
        + "(format_field_name __code_style field.name):" + x10 + ","
        + "(format_field_name __code_style field.name):" + x13 + ","
        + "(format_field_name __code_style field.name):" + x132 + ","
        + "(format_field_name __code_style field.name):" + x14 + ","
        + "(format_field_name __code_style field.name):" + x15 + ","
        + "(format_field_name __code_style field.name):" + v2 + ","
        + "(format_field_name __code_style field.name):" + t1 + ","
        + "(format_field_name __code_style field.name):" + k1 + ","
        + "(format_field_name __code_style field.name):" + k2 + ","
        + "(format_field_name __code_style field.name):" + k8 + ","
        + "(format_field_name __code_style field.name):" + k9 + ","
        + "(format_field_name __code_style field.name):" + k10 + ","
        + "(format_field_name __code_style field.name):" + k11 + ","
        + "}";
    }
}

