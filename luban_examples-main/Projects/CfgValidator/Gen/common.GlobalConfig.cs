
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

using Luban;
using System.Text.Json;


namespace cfg.common
{
public sealed partial class GlobalConfig : Luban.BeanBase
{
    public GlobalConfig(JsonElement _buf) 
    {
        BagCapacity = _buf.GetProperty("bag_capacity").GetInt32();
        BagCapacitySpecial = _buf.GetProperty("bag_capacity_special").GetInt32();
        BagTempExpendableCapacity = _buf.GetProperty("bag_temp_expendable_capacity").GetInt32();
        BagTempToolCapacity = _buf.GetProperty("bag_temp_tool_capacity").GetInt32();
        BagInitCapacity = _buf.GetProperty("bag_init_capacity").GetInt32();
        QuickBagCapacity = _buf.GetProperty("quick_bag_capacity").GetInt32();
        ClothBagCapacity = _buf.GetProperty("cloth_bag_capacity").GetInt32();
        ClothBagInitCapacity = _buf.GetProperty("cloth_bag_init_capacity").GetInt32();
        ClothBagCapacitySpecial = _buf.GetProperty("cloth_bag_capacity_special").GetInt32();
        {if (_buf.TryGetProperty("bag_init_items_drop_id", out var _j) && _j.ValueKind != JsonValueKind.Null) { BagInitItemsDropId = _j.GetInt32(); } else { BagInitItemsDropId = null; } }
        MailBoxCapacity = _buf.GetProperty("mail_box_capacity").GetInt32();
        DamageParamC = _buf.GetProperty("damage_param_c").GetSingle();
        DamageParamE = _buf.GetProperty("damage_param_e").GetSingle();
        DamageParamF = _buf.GetProperty("damage_param_f").GetSingle();
        DamageParamD = _buf.GetProperty("damage_param_d").GetSingle();
        RoleSpeed = _buf.GetProperty("role_speed").GetSingle();
        MonsterSpeed = _buf.GetProperty("monster_speed").GetSingle();
        InitEnergy = _buf.GetProperty("init_energy").GetInt32();
        InitViality = _buf.GetProperty("init_viality").GetInt32();
        MaxViality = _buf.GetProperty("max_viality").GetInt32();
        PerVialityRecoveryTime = _buf.GetProperty("per_viality_recovery_time").GetInt32();
    }

    public static GlobalConfig DeserializeGlobalConfig(JsonElement _buf)
    {
        return new common.GlobalConfig(_buf);
    }

    /// <summary>
    /// 
    /// </summary>
    public readonly int BagCapacity;
    public readonly int BagCapacitySpecial;
    public readonly int BagTempExpendableCapacity;
    public readonly int BagTempToolCapacity;
    public readonly int BagInitCapacity;
    public readonly int QuickBagCapacity;
    public readonly int ClothBagCapacity;
    public readonly int ClothBagInitCapacity;
    public readonly int ClothBagCapacitySpecial;
    public readonly int? BagInitItemsDropId;
    public readonly int MailBoxCapacity;
    public readonly float DamageParamC;
    public readonly float DamageParamE;
    public readonly float DamageParamF;
    public readonly float DamageParamD;
    public readonly float RoleSpeed;
    public readonly float MonsterSpeed;
    public readonly int InitEnergy;
    public readonly int InitViality;
    public readonly int MaxViality;
    public readonly int PerVialityRecoveryTime;
   
    public const int __ID__ = -848234488;
    public override int GetTypeId() => __ID__;

    public  void ResolveRef(Tables tables)
    {
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
    }

    public override string ToString()
    {
        return "{ "
        + "bagCapacity:" + BagCapacity + ","
        + "bagCapacitySpecial:" + BagCapacitySpecial + ","
        + "bagTempExpendableCapacity:" + BagTempExpendableCapacity + ","
        + "bagTempToolCapacity:" + BagTempToolCapacity + ","
        + "bagInitCapacity:" + BagInitCapacity + ","
        + "quickBagCapacity:" + QuickBagCapacity + ","
        + "clothBagCapacity:" + ClothBagCapacity + ","
        + "clothBagInitCapacity:" + ClothBagInitCapacity + ","
        + "clothBagCapacitySpecial:" + ClothBagCapacitySpecial + ","
        + "bagInitItemsDropId:" + BagInitItemsDropId + ","
        + "mailBoxCapacity:" + MailBoxCapacity + ","
        + "damageParamC:" + DamageParamC + ","
        + "damageParamE:" + DamageParamE + ","
        + "damageParamF:" + DamageParamF + ","
        + "damageParamD:" + DamageParamD + ","
        + "roleSpeed:" + RoleSpeed + ","
        + "monsterSpeed:" + MonsterSpeed + ","
        + "initEnergy:" + InitEnergy + ","
        + "initViality:" + InitViality + ","
        + "maxViality:" + MaxViality + ","
        + "perVialityRecoveryTime:" + PerVialityRecoveryTime + ","
        + "}";
    }
}

}
