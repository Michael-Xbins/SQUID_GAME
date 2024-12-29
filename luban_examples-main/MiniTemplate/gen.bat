set WORKSPACE=..

set LUBAN_DLL=%WORKSPACE%\Tools\Luban\Luban.dll
set CONF_ROOT=.
set TARGET_DIR=..\..\application\config

dotnet %LUBAN_DLL% ^
    -t all ^
    -c go-json ^
    -d json  ^
    --schemaPath %CONF_ROOT%\Defines\__root__.xml ^
    -x inputDataDir=%CONF_ROOT%\Datas ^
	-x outputCodeDir=%TARGET_DIR%\gen ^
    -x outputDataDir=%TARGET_DIR%\output_json ^
    -x pathValidator.rootDir=%WORKSPACE%\Projects\Csharp_Unity_bin ^
    -x l10n.textProviderFile=*@%WORKSPACE%\DesignerConfigs\Datas\l10n\texts.json ^
    -x lubanGoModule=demo/luban


pause