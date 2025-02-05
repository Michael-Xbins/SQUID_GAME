
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

using Luban;
using System.Text.Json;


namespace cfg.item
{
/// <summary>
/// 道具
/// </summary>
public sealed partial class Item : Luban.BeanBase
{
    public Item(JsonElement _buf) 
    {
        Id = _buf.GetProperty("id").GetInt32();
        Name = _buf.GetProperty("name").GetString();
        MajorType = (item.EMajorType)_buf.GetProperty("major_type").GetInt32();
        MinorType = (item.EMinorType)_buf.GetProperty("minor_type").GetInt32();
        MaxPileNum = _buf.GetProperty("max_pile_num").GetInt32();
        Quality = (item.EItemQuality)_buf.GetProperty("quality").GetInt32();
        Icon = _buf.GetProperty("icon").GetString();
        IconBackgroud = _buf.GetProperty("icon_backgroud").GetString();
        IconMask = _buf.GetProperty("icon_mask").GetString();
        Desc = _buf.GetProperty("desc").GetString();
        ShowOrder = _buf.GetProperty("show_order").GetInt32();
    }

    public static Item DeserializeItem(JsonElement _buf)
    {
        return new item.Item(_buf);
    }

    /// <summary>
    /// 
    /// </summary>
    public readonly int Id;
    public readonly string Name;
    public readonly item.EMajorType MajorType;
    public readonly item.EMinorType MinorType;
    public readonly int MaxPileNum;
    public readonly item.EItemQuality Quality;
    public readonly string Icon;
    public readonly string IconBackgroud;
    public readonly string IconMask;
    public readonly string Desc;
    public readonly int ShowOrder;
   
    public const int __ID__ = 2107285806;
    public override int GetTypeId() => __ID__;

    public  void ResolveRef(Tables tables)
    {
        
        
        
        
        
        
        
        
        
        
        
    }

    public override string ToString()
    {
        return "{ "
        + "id:" + Id + ","
        + "name:" + Name + ","
        + "majorType:" + MajorType + ","
        + "minorType:" + MinorType + ","
        + "maxPileNum:" + MaxPileNum + ","
        + "quality:" + Quality + ","
        + "icon:" + Icon + ","
        + "iconBackgroud:" + IconBackgroud + ","
        + "iconMask:" + IconMask + ","
        + "desc:" + Desc + ","
        + "showOrder:" + ShowOrder + ","
        + "}";
    }
}

}
