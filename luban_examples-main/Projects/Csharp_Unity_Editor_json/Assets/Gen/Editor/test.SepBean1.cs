
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

using System.Collections.Generic;
using SimpleJSON;
using Luban;

namespace editor.cfg.test
{

public sealed class SepBean1 :  Luban.EditorBeanBase 
{
    public SepBean1()
    {
            c = "";
    }

    public override void LoadJson(SimpleJSON.JSONObject _json)
    {
        { 
            var _fieldJson = _json["a"];
            if (_fieldJson != null)
            {
                if(!_fieldJson.IsNumber) { throw new SerializationException(); }  a = _fieldJson;
            }
        }
        
        { 
            var _fieldJson = _json["b"];
            if (_fieldJson != null)
            {
                if(!_fieldJson.IsNumber) { throw new SerializationException(); }  b = _fieldJson;
            }
        }
        
        { 
            var _fieldJson = _json["c"];
            if (_fieldJson != null)
            {
                if(!_fieldJson.IsString) { throw new SerializationException(); }  c = _fieldJson;
            }
        }
        
    }

    public override void SaveJson(SimpleJSON.JSONObject _json)
    {
        {
            _json["a"] = new JSONNumber(a);
        }
        {
            _json["b"] = new JSONNumber(b);
        }
        {

            if (c == null) { throw new System.ArgumentNullException(); }
            _json["c"] = new JSONString(c);
        }
    }

    public static SepBean1 LoadJsonSepBean1(SimpleJSON.JSONNode _json)
    {
        SepBean1 obj = new test.SepBean1();
        obj.LoadJson((SimpleJSON.JSONObject)_json);
        return obj;
    }
        
    public static void SaveJsonSepBean1(SepBean1 _obj, SimpleJSON.JSONNode _json)
    {
        _obj.SaveJson((SimpleJSON.JSONObject)_json);
    }

    public int a;

    public int b;

    public string c;

}

}
