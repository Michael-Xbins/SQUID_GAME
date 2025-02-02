
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg.test;

import luban.*;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;


public final class TestScriptableObject extends AbstractBean {
    public TestScriptableObject(JsonObject _buf) { 
        id = _buf.get("id").getAsInt();
        desc = _buf.get("desc").getAsString();
        rate = _buf.get("rate").getAsFloat();
        num = _buf.get("num").getAsInt();
        v2 = cfg.vector2.deserialize(_buf.get("v2").getAsJsonObject());
        v3 = cfg.vector3.deserialize(_buf.get("v3").getAsJsonObject());
        v4 = cfg.vector4.deserialize(_buf.get("v4").getAsJsonObject());
    }

    public static TestScriptableObject deserialize(JsonObject _buf) {
            return new cfg.test.TestScriptableObject(_buf);
    }

    public final int id;
    public final String desc;
    public final float rate;
    public final int num;
    public final cfg.vector2 v2;
    public final cfg.vector3 v3;
    public final cfg.vector4 v4;

    public static final int __ID__ = -1896814350;
    
    @Override
    public int getTypeId() { return __ID__; }

    @Override
    public String toString() {
        return "{ "
        + "(format_field_name __code_style field.name):" + id + ","
        + "(format_field_name __code_style field.name):" + desc + ","
        + "(format_field_name __code_style field.name):" + rate + ","
        + "(format_field_name __code_style field.name):" + num + ","
        + "(format_field_name __code_style field.name):" + v2 + ","
        + "(format_field_name __code_style field.name):" + v3 + ","
        + "(format_field_name __code_style field.name):" + v4 + ","
        + "}";
    }
}

