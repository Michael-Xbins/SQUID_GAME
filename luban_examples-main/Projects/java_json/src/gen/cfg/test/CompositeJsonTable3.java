
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


public final class CompositeJsonTable3 extends AbstractBean {
    public CompositeJsonTable3(JsonObject _buf) { 
        a = _buf.get("a").getAsInt();
        b = _buf.get("b").getAsInt();
    }

    public static CompositeJsonTable3 deserialize(JsonObject _buf) {
            return new cfg.test.CompositeJsonTable3(_buf);
    }

    public final int a;
    public final int b;

    public static final int __ID__ = 1566207896;
    
    @Override
    public int getTypeId() { return __ID__; }

    @Override
    public String toString() {
        return "{ "
        + "(format_field_name __code_style field.name):" + a + ","
        + "(format_field_name __code_style field.name):" + b + ","
        + "}";
    }
}

