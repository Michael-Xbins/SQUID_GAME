
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


public final class Item extends cfg.test.ItemBase {
    public Item(JsonObject _buf) { 
        super(_buf);
        num = _buf.get("num").getAsInt();
        price = _buf.get("price").getAsInt();
    }

    public static Item deserialize(JsonObject _buf) {
            return new cfg.test.Item(_buf);
    }

    public final int num;
    public final int price;

    public static final int __ID__ = -1226641649;
    
    @Override
    public int getTypeId() { return __ID__; }

    @Override
    public String toString() {
        return "{ "
        + "(format_field_name __code_style field.name):" + id + ","
        + "(format_field_name __code_style field.name):" + name + ","
        + "(format_field_name __code_style field.name):" + desc + ","
        + "(format_field_name __code_style field.name):" + num + ","
        + "(format_field_name __code_style field.name):" + price + ","
        + "}";
    }
}

