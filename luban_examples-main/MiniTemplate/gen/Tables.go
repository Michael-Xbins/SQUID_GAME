
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;

type JsonLoader func(string) ([]map[string]interface{}, error)

type Tables struct {
    TBApp *AppTBApp
    TBWood *WoodTBWood
    TBCompete *CompeteTBCompete
    TBGlass *GlassTBGlass
}

func NewTables(loader JsonLoader) (*Tables, error) {
    var err error
    var buf []map[string]interface{}

    tables := &Tables{}
    if buf, err = loader("app_tbapp") ; err != nil {
        return nil, err
    }
    if tables.TBApp, err = NewAppTBApp(buf) ; err != nil {
        return nil, err
    }
    if buf, err = loader("wood_tbwood") ; err != nil {
        return nil, err
    }
    if tables.TBWood, err = NewWoodTBWood(buf) ; err != nil {
        return nil, err
    }
    if buf, err = loader("compete_tbcompete") ; err != nil {
        return nil, err
    }
    if tables.TBCompete, err = NewCompeteTBCompete(buf) ; err != nil {
        return nil, err
    }
    if buf, err = loader("glass_tbglass") ; err != nil {
        return nil, err
    }
    if tables.TBGlass, err = NewGlassTBGlass(buf) ; err != nil {
        return nil, err
    }
    return tables, nil
}


