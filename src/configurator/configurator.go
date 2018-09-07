package configurator

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

const (
	CfgCommonPrintMemoryInfo    string = "common.PrintMemoryInfo"
	CfgCommonPrintAperturesInfo string = "common.PrintAperturesInfo"
	CfgCommonPrintRegionsInfo   string = "common.PrintRegionsInfo"
	CfgCommonPrintStatistic   string = "common.PrintStatistic"
	CfgParserSaveIntermediate string = "parser.SaveIntermediate"

)

func SetDefaults(v *viper.Viper) {
	v.SetConfigName("config") // no need to include file extension
	v.AddConfigPath(".")      // set the path of your config file
	v.SetConfigType("toml")

	// diagnostic messages
	v.SetDefault(CfgCommonPrintMemoryInfo, true)
	v.SetDefault(CfgCommonPrintAperturesInfo, true)
	v.SetDefault(CfgCommonPrintRegionsInfo, true)
	v.SetDefault(CfgCommonPrintStatistic, true)

	//
	v.SetDefault(CfgParserSaveIntermediate, true)
	v.SetDefault("renderer.GeneratePNG", true)
	v.SetDefault("renderer.OutFile", "out.png")

	//
	v.SetDefault("pcb.xOrigin", 0)
	v.SetDefault("pcb.yOrigin", 0)

	//
	v.SetDefault("renderer.CanvasWidth", 297)
	v.SetDefault("renderer.CanvasHeight", 210)
	//
	v.SetDefault("plotter.PenSizes", []float64{0.07, 0.07, 0.07, 0.00})
	v.SetDefault("plotter.xRes", 0.025)
	v.SetDefault("plotter.yRes", 0.025)

}

func ProcessConfigFile(v *viper.Viper) error {
	return v.ReadInConfig()
	return errors.New("configuration file error. Using defaults")
}

func DiagnosticAllCfgPrint(v *viper.Viper) {
	c := v.AllSettings()
	for key, data := range c {
		fmt.Println(key, ":", data)
	}
	fmt.Println()
	return
}
